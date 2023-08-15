package handlers

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"net/http"
	"net/url"
	"strings"
)

func Subscription(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		p, _ := url.QueryUnescape(c.Request().RequestURI[14:])
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

		authPart := b64.StdEncoding.EncodeToString([]byte(key.Cipher + ":" + key.Secret))

		var lines []string
		for _, s := range servers {
			lines = append(lines, fmt.Sprintf(
				"ss://%s@%s:%d/?outline=1#%s:%d",
				authPart, s.ShadowsocksHost, s.ShadowsocksPort, s.ShadowsocksHost, s.ShadowsocksPort,
			))
		}

		return c.String(http.StatusOK, b64.StdEncoding.EncodeToString([]byte(strings.Join(lines, "\n"))))
	}
}
