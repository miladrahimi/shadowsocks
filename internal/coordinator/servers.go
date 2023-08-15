package coordinator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func (c *Coordinator) syncServers(reconfigure bool) {
	c.Logger.Debug("syncing servers with the prometheus service")

	servers := map[string]string{c.CurrentServer().Id: fmt.Sprintf("127.0.0.1:%d", c.CurrentServer().HttpPort)}
	for _, s := range c.Database.ServerTable.Servers {
		servers[s.Id] = fmt.Sprintf("%s:%d", s.HttpHost, s.HttpPort)
	}

	if err := c.Prometheus.Update(servers); err != nil {
		c.Logger.Fatal("cannot set servers in the prometheus service", zap.Error(err))
	}

	if reconfigure {
		c.Prometheus.Reload()
	}

	go c.pushServers()
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

func (c *Coordinator) updateServerStatus(s *database.Server, newStatus string) {
	s.Status = newStatus
	if _, err := c.Database.ServerTable.Update(*s); err != nil {
		c.Logger.Error("cannot update server status", zap.String("server", s.Id), zap.Error(err))
	}
}

func (c *Coordinator) pullServers() {
	for _, s := range c.Database.ServerTable.Servers {
		go c.pullServer(s)
	}
}

func (c *Coordinator) pullServer(s *database.Server) {
	url := fmt.Sprintf("http://%s:%d/v1/settings", s.HttpHost, s.HttpPort)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.Logger.Error("cannot create server pull request", zap.String("url", url), zap.Error(err))
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}

	request.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Add(echo.HeaderAuthorization, "Bearer "+s.ApiToken)

	response, err := c.Http.Do(request)
	if err != nil {
		c.Logger.Error("cannot pull server", zap.String("url", url), zap.Error(err))
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusUnauthorized {
		c.updateServerStatus(s, database.ServerStatusUnauthorized)
		return
	}

	if response.StatusCode != http.StatusOK {
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.Logger.Error("cannot read pulled server", zap.String("url", url), zap.Error(err))
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}

	var settings database.SettingTable
	if err = json.Unmarshal(body, &settings); err != nil {
		c.Logger.Error(
			"cannot unmarshall pulled server", zap.String("url", url),
			zap.Error(err), zap.String("body", string(body)),
		)
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}

	s.Status = database.ServerStatusActive
	s.ShadowsocksEnabled = settings.ShadowsocksEnabled
	s.ShadowsocksHost = settings.ShadowsocksHost
	s.ShadowsocksPort = settings.ShadowsocksPort

	if _, err = c.Database.ServerTable.Update(*s); err != nil {
		c.Logger.Error("cannot update server", zap.String("server", s.Id), zap.Error(err))
	}
}

func (c *Coordinator) pushServers() {
	for _, s := range c.Database.ServerTable.Servers {
		go c.pushServer(s)
	}
}

func (c *Coordinator) pushServer(s *database.Server) {
	url := fmt.Sprintf("http://%s:%d/v1/keys/refill", s.HttpHost, s.HttpPort)
	c.Logger.Debug("pushing keys to server...", zap.String("url", url))

	body, err := json.Marshal(c.Database.KeyTable.Keys)
	if err != nil {
		c.Logger.Fatal("cannot marshal database.keys", zap.Error(err))
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		c.Logger.Error("cannot create push request", zap.String("url", url), zap.Error(err))
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}

	request.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request.Header.Add(echo.HeaderAuthorization, "Bearer "+s.ApiToken)

	response, err := c.Http.Do(request)
	if err != nil {
		c.Logger.Error("cannot push keys to server", zap.String("url", url), zap.Error(err))
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusUnauthorized {
		c.updateServerStatus(s, database.ServerStatusUnauthorized)
		return
	}

	if response.StatusCode != http.StatusNoContent {
		c.updateServerStatus(s, database.ServerStatusUnavailable)
		return
	}

	s.Status = database.ServerStatusActive
	s.SyncedAt = time.Now().Unix()

	if _, err = c.Database.ServerTable.Update(*s); err != nil {
		c.Logger.Error("cannot update server", zap.String("server", s.Id), zap.Error(err))
	}
}
