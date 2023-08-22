package server

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/miladrahimi/shadowsocks/internal/config"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/http/handlers"
	"github.com/miladrahimi/shadowsocks/internal/http/handlers/v1"
	internalMw "github.com/miladrahimi/shadowsocks/internal/http/middleware"
	"github.com/miladrahimi/shadowsocks/internal/http/validator"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Server struct {
	Engine      *echo.Echo
	config      *config.Config
	logger      *zap.Logger
	coordinator *coordinator.Coordinator
}

func New(config *config.Config, logger *zap.Logger, coordinator *coordinator.Coordinator) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.New()

	return &Server{Engine: e, config: config, logger: logger, coordinator: coordinator}
}

func (s *Server) Run() {
	s.Engine.Use(middleware.CORS())
	s.Engine.Use(internalMw.Logger(s.logger))

	s.Engine.Static("/", "web")

	s.Engine.GET("/metrics", handlers.Metrics(s.coordinator))
	s.Engine.GET("/ssconf/*", handlers.SSConf(s.coordinator))
	s.Engine.GET("/subscription/*", handlers.Subscription(s.coordinator))
	s.Engine.GET("/public", handlers.Public(s.coordinator))
	s.Engine.GET("/profile", handlers.Profile(s.coordinator))

	g1 := s.Engine.Group("/v1")
	g1.POST("/sign-in", v1.SignIn(s.coordinator))
	g1.GET("/profile", v1.ProfileShow(s.coordinator))
	g1.POST("/profile/reset", v1.ProfileReset(s.coordinator))

	g2 := s.Engine.Group("/v1")
	g2.Use(internalMw.Authorize(s.coordinator.Database))
	s.Engine.GET("/health", v1.Health())
	g2.GET("/settings", v1.SettingsShow(s.config, s.coordinator))
	g2.POST("/settings", v1.SettingsUpdate(s.coordinator))
	g2.GET("/servers", v1.ServersIndex(s.coordinator))
	g2.POST("/servers", v1.ServersStore(s.coordinator))
	g2.PUT("/servers", v1.ServersUpdate(s.coordinator))
	g2.DELETE("/servers/:id", v1.ServersDelete(s.coordinator))
	g2.GET("/keys", v1.KeysIndex(s.coordinator))
	g2.POST("/keys", v1.KeysStore(s.coordinator))
	g2.PUT("/keys", v1.KeysUpdate(s.coordinator))
	g2.DELETE("/keys/:id", v1.KeysDelete(s.coordinator))
	g2.PATCH("/keys/:id/empty", v1.KeysEmpty(s.coordinator))
	g2.POST("/keys/fill", v1.KeysFill(s.coordinator))

	address := fmt.Sprintf("%s:%d", s.config.HttpServer.Host, s.config.HttpServer.Port)
	if err := s.Engine.Start(address); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("cannot start the http server", zap.String("address", address), zap.Error(err))
	}
}

func (s *Server) Shutdown() {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Engine.Shutdown(c); err != nil {
		s.logger.Error("cannot close the http server", zap.Error(err))
	} else {
		s.logger.Debug("http server closed successfully")
	}
}
