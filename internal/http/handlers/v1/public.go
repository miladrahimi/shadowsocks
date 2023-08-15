package v1

import (
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"github.com/miladrahimi/shadowsocks/pkg/utils"
	"net/http"
	"strings"
)

type PublicResponse struct {
	coordinator.KeyMetric
	KeyResponse
}

func PublicProfile(cdr *coordinator.Coordinator) echo.HandlerFunc {
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

		var r PublicResponse
		r.KeyResponse.Key = *key
		r.GenerateLinks(cdr)
		if m, found := cdr.KeyMetrics[key.Id]; found {
			r.Quota = int64(float64(r.Quota) * cdr.Database.SettingTable.TrafficRatio)
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

func PublicReset(cdr *coordinator.Coordinator) echo.HandlerFunc {
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

		key.Secret = utils.Token()
		newKey, err := cdr.Database.KeyTable.Update(*key)
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

		kr := &KeyResponse{Key: *newKey}
		kr.GenerateLinks(cdr)

		return c.JSON(http.StatusOK, kr)
	}
}
