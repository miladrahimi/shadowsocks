package handlers

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"net/http"
	"strings"
)

func Public(cdr *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var query, err = base64.StdEncoding.DecodeString(c.QueryParam("k"))
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		parts := strings.Split(string(query), ":")
		if len(parts) != 2 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		var key *database.Key
		for _, k := range cdr.Database.KeyTable.Keys {
			if k.Cipher == parts[0] && k.Secret == parts[1] {
				key = k
			}
		}
		if key == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		return c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("/profile?c=%s", key.Code))
	}
}
