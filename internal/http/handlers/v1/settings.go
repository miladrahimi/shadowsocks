package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/config"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"net/http"
)

type SettingsResponse struct {
	database.SettingTable
	HttpPort int `json:"http_port"`
}

func SettingsShow(cfg *config.Config, coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, SettingsResponse{
			SettingTable: *coordinator.Database.SettingTable,
			HttpPort:     cfg.HttpServer.Port,
		})
	}
}

func SettingsUpdate(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r database.SettingTable
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		coordinator.Database.SettingTable.ExternalHttps = r.ExternalHttps
		coordinator.Database.SettingTable.ExternalHttp = r.ExternalHttp
		coordinator.Database.SettingTable.ShadowsocksHost = r.ShadowsocksHost
		coordinator.Database.SettingTable.ShadowsocksPort = r.ShadowsocksPort
		coordinator.Database.SettingTable.ShadowsocksEnabled = r.ShadowsocksEnabled
		coordinator.Database.SettingTable.ApiToken = r.ApiToken
		coordinator.Database.SettingTable.AdminPassword = r.AdminPassword
		coordinator.Database.SettingTable.TrafficRatio = r.TrafficRatio

		if err := coordinator.Database.SettingTable.Save(); err != nil {
			if _, ok := err.(database.DataError); ok {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": err.Error(),
				})
			}
			return err
		}

		go coordinator.Sync()

		return c.JSON(http.StatusOK, coordinator.Database.SettingTable)
	}
}
