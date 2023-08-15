package app

import (
	"context"
	"github.com/miladrahimi/shadowsocks/internal/config"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"github.com/miladrahimi/shadowsocks/internal/http/client"
	"github.com/miladrahimi/shadowsocks/internal/http/server"
	"github.com/miladrahimi/shadowsocks/internal/logger"
	"github.com/miladrahimi/shadowsocks/pkg/prometheus"
	"github.com/miladrahimi/shadowsocks/pkg/shadowsocks"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const shadowsocksKeysPath = "storage/shadowsocks/keys.yml"
const prometheusConfigPath = "storage/prometheus/configs/prometheus.yml"

var shadowsocksBinaryPaths = map[string]string{
	"darwin": "third_party/outline-macos-arm64/outline-ss-server",
	"linux":  "third_party/outline-linux-x86_64/outline-ss-server",
}

// App integrates the modules to serve.
type App struct {
	Context     context.Context
	Config      *config.Config
	Logger      *logger.Logger
	HttpClient  *http.Client
	HttpServer  *server.Server
	Prometheus  *prometheus.Prometheus
	Shadowsocks *shadowsocks.Shadowsocks
	Database    *database.Database
	Coordinator *coordinator.Coordinator
}

// New creates an instance of the application.
func New() (app *App, err error) {
	app = &App{}

	if app.Config, err = config.New(); err != nil {
		return app, err
	}
	if app.Logger, err = logger.New(app.Config); err != nil {
		return app, err
	}
	app.Logger.Engine.Debug("config and logger modules initialized")

	app.HttpClient = client.New(app.Config)
	app.Logger.Engine.Debug("http client initialized")

	if app.Database, err = database.New(); err != nil {
		return app, err
	}
	app.Logger.Engine.Debug("database initialized and loaded")

	app.Shadowsocks = shadowsocks.New(app.Logger.Engine, shadowsocksKeysPath, shadowsocksBinaryPaths)
	app.Logger.Engine.Debug("shadowsocks initialized")

	app.Prometheus = prometheus.New(
		app.Logger.Engine, app.HttpClient, prometheusConfigPath, app.Config.Prometheus.Host, app.Config.Prometheus.Port,
	)
	app.Logger.Engine.Debug("prometheus initialized")

	app.Coordinator = coordinator.New(
		app.Config, app.Logger.Engine, app.HttpClient, app.Prometheus, app.Database, app.Shadowsocks,
	)
	app.Logger.Engine.Debug("coordinator initialized")

	app.HttpServer = server.New(app.Config, app.Logger.Engine, app.Coordinator)
	app.Logger.Engine.Debug("http server initialized")

	app.setupSignalListener()

	return app, nil
}

// setupSignalListener sets up a listener to signals from os.
func (a *App) setupSignalListener() {
	var cancel context.CancelFunc
	a.Context, cancel = context.WithCancel(context.Background())

	// Listen to SIGTERM
	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		s := <-signalChannel
		a.Logger.Engine.Info("system call", zap.String("signal", s.String()))

		cancel()
	}()

	// Listen to SIGHUP
	go func() {
		for {
			signalChannel := make(chan os.Signal, 2)
			signal.Notify(signalChannel, syscall.SIGHUP)

			s := <-signalChannel
			a.Logger.Engine.Info("system call", zap.String("signal", s.String()))

			a.Shadowsocks.Reconfigure()
		}
	}()
}

// Wait avoid dying app and shut it down gracefully on exit signals.
func (a *App) Wait() {
	<-a.Context.Done()
}

func (a *App) Shutdown() {
	if a.Shadowsocks != nil {
		a.Shadowsocks.Shutdown()
	}
	if a.HttpServer != nil {
		a.HttpServer.Shutdown()
	}
	if a.Logger != nil {
		a.Logger.Shutdown()
	}
}
