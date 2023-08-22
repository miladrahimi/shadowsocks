package v1

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"net/http"
	"strings"
)

type ProfileResponse struct {
	KeyResponse
	coordinator.KeyMetric
	SSCONF       string   `json:"ssconf"`
	Subscription string   `json:"subscription"`
	SSKeys       []string `json:"ss_keys"`
}

func ProfileShow(cdr *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		key, err := cdr.Database.KeyTable.FindByCode(c.QueryParam("c"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "Cannot update the database.",
			})
		}
		if key == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		var r ProfileResponse
		r.KeyResponse.Key = key
		r.Quota = int64(float64(r.Quota) * cdr.Database.SettingTable.TrafficRatio)

		auth := base64.StdEncoding.EncodeToString([]byte(r.Cipher + ":" + r.Secret))
		settings := cdr.Database.SettingTable

		if settings.ExternalHttps != "" {
			url := strings.Replace(settings.ExternalHttps, "https://", "ssconf://", 1)
			r.SSCONF = fmt.Sprintf("%s/ssconf/%s.json#%s", url, auth, r.Name)
		}

		if settings.ExternalHttp != "" {
			r.Subscription = fmt.Sprintf("%s/subscription/%s#%s", settings.ExternalHttp, auth, r.Name)
		}

		for _, s := range append(cdr.Database.ServerTable.Servers, cdr.CurrentServer()) {
			if s.ShadowsocksEnabled && s.Status == database.ServerStatusActive {
				r.SSKeys = append(r.SSKeys, fmt.Sprintf(
					"ss://%s@%s:%d/?outline=1#%s", auth, s.ShadowsocksHost, s.ShadowsocksPort, r.Name,
				))
			}
		}

		if m, found := cdr.KeyMetrics[key.Id]; found {
			r.KeyMetric = coordinator.KeyMetric{
				Id:      m.Id,
				DownTcp: int64(float64(m.DownTcp)*cdr.Database.SettingTable.TrafficRatio) / 1000000,
				DownUdp: int64(float64(m.DownUdp)*cdr.Database.SettingTable.TrafficRatio) / 1000000,
				UpTcp:   int64(float64(m.UpTcp)*cdr.Database.SettingTable.TrafficRatio) / 1000000,
				UpUdp:   int64(float64(m.UpUdp)*cdr.Database.SettingTable.TrafficRatio) / 1000000,
				Total:   int64(float64(m.Total)*cdr.Database.SettingTable.TrafficRatio) / 1000000,
			}
		}

		return c.JSON(http.StatusOK, r)
	}
}

func ProfileReset(cdr *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		key, err := cdr.Database.KeyTable.FindByCode(c.QueryParam("c"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "Cannot update the database.",
			})
		}
		if key == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		key.Secret = random.String(16)
		key, err = cdr.Database.KeyTable.Update(*key)
		if err != nil {
			if _, ok := err.(database.DataError); ok {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": err.Error(),
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Internal error.",
			})
		}
		cdr.Sync()

		return c.JSON(http.StatusOK, key)
	}
}
