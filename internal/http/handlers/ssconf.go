package handlers

import (
	b64 "encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"golang.org/x/exp/rand"
	"net/http"
	"net/url"
	"strings"
)

func SSConf(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		jIndex := strings.Index(c.Request().RequestURI, ".json")
		p, _ := url.QueryUnescape(c.Request().RequestURI[8:jIndex])
		var auth, err = b64.StdEncoding.DecodeString(p)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		parts := strings.Split(string(auth), ":")
		if len(parts) != 2 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		var key *database.Key
		for _, k := range coordinator.Database.KeyTable.Keys {
			if k.Cipher == parts[0] && k.Secret == parts[1] {
				key = k
			}
		}
		if key == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		var servers []*database.Server
		if coordinator.CurrentServer().ShadowsocksEnabled {
			servers = append(servers, coordinator.CurrentServer())
		}
		for _, s := range coordinator.Database.ServerTable.Servers {
			if s.Status == database.ServerStatusActive {
				servers = append(servers, s)
			}
		}

		if len(servers) == 0 {
			return c.JSON(http.StatusNotFound, map[string]interface{}{})
		}

		randomServerIndex := rand.Intn(len(servers))
		server := servers[randomServerIndex]

		return c.JSON(http.StatusOK, map[string]interface{}{
			"server":      server.ShadowsocksHost,
			"server_port": server.ShadowsocksPort,
			"password":    key.Secret,
			"method":      key.Cipher,
		})
	}
}
