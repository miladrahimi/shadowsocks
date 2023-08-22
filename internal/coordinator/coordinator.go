package coordinator

import (
	"fmt"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/shadowsocks/internal/config"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"github.com/miladrahimi/shadowsocks/pkg/prometheus"
	"github.com/miladrahimi/shadowsocks/pkg/shadowsocks"
	"github.com/miladrahimi/shadowsocks/pkg/utils"
	"go.uber.org/zap"
	"net/http"
)

type Coordinator struct {
	Http          *http.Client
	Logger        *zap.Logger
	Config        *config.Config
	Prometheus    *prometheus.Prometheus
	Shadowsocks   *shadowsocks.Shadowsocks
	Database      *database.Database
	MetricsPort   int
	ServerMetrics map[string]*ServerMetric
	KeyMetrics    map[string]*KeyMetric
	SyncedAt      int64
}

func (c *Coordinator) Run() {
	c.initSettings()
	c.initMetricsPort()
	c.syncShadowsocks(false)
	c.syncServers(false)
	go c.Shadowsocks.Run(c.MetricsPort)
	go c.Prometheus.Reload()
	go c.startWorkers()
}

func (c *Coordinator) Sync() {
	c.syncShadowsocks(true)
	c.syncServers(true)
}

func (c *Coordinator) initSettings() {
	if c.Database.SettingTable.ApiToken == "api-token-secret" {
		c.Database.SettingTable.ApiToken = random.String(32)
	}

	if c.Database.SettingTable.ShadowsocksPort == 1 {
		var err error
		if c.Database.SettingTable.ShadowsocksPort, err = utils.FreePort(); err != nil {
			c.Logger.Fatal("cannot find a free port for the shadowsocks server", zap.Error(err))
		}
	}

	if c.Database.SettingTable.ExternalHttp == "http://empty" {
		c.Database.SettingTable.ExternalHttp = fmt.Sprintf("http://127.0.0.1:%d", c.Config.HttpServer.Port)
	}

	if err := c.Database.SettingTable.Save(); err != nil {
		c.Logger.Fatal("cannot save settings", zap.Error(err))
	}
}

func (c *Coordinator) initMetricsPort() {
	var err error
	if c.MetricsPort, err = utils.FreePort(); err != nil {
		c.Logger.Fatal("cannot find a free port for the shadowsocks metrics", zap.Error(err))
	}
}

func (c *Coordinator) CurrentServer() *database.Server {
	return &database.Server{
		Id:                 "s-0",
		Status:             database.ServerStatusActive,
		HttpHost:           "127.0.0.1",
		HttpPort:           c.Config.HttpServer.Port,
		ShadowsocksEnabled: c.Database.SettingTable.ShadowsocksEnabled,
		ShadowsocksHost:    c.Database.SettingTable.ShadowsocksHost,
		ShadowsocksPort:    c.Database.SettingTable.ShadowsocksPort,
		ApiToken:           c.Database.SettingTable.ApiToken,
		SyncedAt:           c.SyncedAt,
	}
}

func New(
	c *config.Config,
	l *zap.Logger,
	hc *http.Client,
	p *prometheus.Prometheus,
	db *database.Database,
	ss *shadowsocks.Shadowsocks,
) *Coordinator {
	return &Coordinator{
		Config:        c,
		Logger:        l,
		Http:          hc,
		Database:      db,
		Prometheus:    p,
		Shadowsocks:   ss,
		ServerMetrics: map[string]*ServerMetric{},
		KeyMetrics:    map[string]*KeyMetric{},
	}
}
