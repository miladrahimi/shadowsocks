package v1

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"net/http"
	"strings"
)

type KeysStoreRequest struct {
	Cipher  string `json:"cipher"`
	Secret  string `json:"secret"`
	Name    string `json:"name"`
	Quota   int64  `json:"quota"`
	Enabled bool   `json:"enabled"`
}

type KeysUpdateRequest struct {
	KeysStoreRequest
	Id string `json:"id"`
}

type KeyResponse struct {
	database.Key
	Used         int64    `json:"used"`
	Public       string   `json:"public"`
	SSCONF       string   `json:"ssconf"`
	Subscription string   `json:"subscription"`
	SSKeys       []string `json:"ss_keys"`
}

func (k *KeyResponse) GenerateLinks(c *coordinator.Coordinator) {
	auth := base64.StdEncoding.EncodeToString([]byte(k.Cipher + ":" + k.Secret))

	if c.Database.SettingTable.ExternalHttps != "" {
		url := strings.Replace(c.Database.SettingTable.ExternalHttps, "https://", "ssconf://", 1)
		k.SSCONF = fmt.Sprintf("%s/ssconf/%s.json#%s", url, auth, k.Name)
		k.Public = fmt.Sprintf("%s/public?k=%s", c.Database.SettingTable.ExternalHttps, auth)
	} else {
		k.Public = fmt.Sprintf("%s/public?k=%s", c.Database.SettingTable.ExternalHttp, auth)
	}

	k.Subscription = fmt.Sprintf("%s/subscription/%s#%s", c.Database.SettingTable.ExternalHttp, auth, k.Name)

	for _, s := range append(c.Database.ServerTable.Servers, c.CurrentServer()) {
		if s.ShadowsocksEnabled && s.Status == database.ServerStatusActive {
			k.SSKeys = append(k.SSKeys, fmt.Sprintf(
				"ss://%s@%s:%d/?outline=1#%s", auth, s.ShadowsocksHost, s.ShadowsocksPort, k.Name,
			))
		}
	}
}

func KeysIndex(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		krs := make([]KeyResponse, 0, len(coordinator.Database.KeyTable.Keys))
		for _, k := range coordinator.Database.KeyTable.Keys {
			kr := &KeyResponse{Key: *k}
			kr.GenerateLinks(coordinator)
			if m, found := coordinator.KeyMetrics[k.Id]; found {
				kr.Used = m.Total / 1000000
			}
			krs = append(krs, *kr)
		}

		return c.JSON(http.StatusOK, krs)
	}
}

func KeysStore(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r KeysStoreRequest
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		key, err := coordinator.Database.KeyTable.Store(database.Key{
			Cipher:  r.Cipher,
			Secret:  r.Secret,
			Name:    r.Name,
			Quota:   r.Quota,
			Enabled: r.Enabled,
		})
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

		go coordinator.Sync()

		return c.JSON(http.StatusCreated, KeyResponse{Key: *key, Used: 0})
	}
}

func KeysUpdate(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r KeysUpdateRequest
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		key, err := coordinator.Database.KeyTable.Update(database.Key{
			Id:      r.Id,
			Cipher:  r.Cipher,
			Secret:  r.Secret,
			Name:    r.Name,
			Quota:   r.Quota,
			Enabled: r.Enabled,
		})
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
		if key == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Key not found.",
			})
		}

		go coordinator.Sync()

		kr := KeyResponse{Key: *key}
		kr.GenerateLinks(coordinator)
		if m, found := coordinator.KeyMetrics[kr.Id]; found {
			kr.Used = m.Total / 1000000
		}

		return c.JSON(http.StatusOK, kr)
	}
}

func KeysReset(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		key, err := coordinator.Database.KeyTable.ReId(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "Cannot update the database.",
			})
		}

		coordinator.Sync()

		return c.JSON(http.StatusOK, key)
	}
}

func KeysDelete(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		err := coordinator.Database.KeyTable.Delete(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "Cannot update the database.",
			})
		}

		go coordinator.Sync()

		return c.NoContent(http.StatusNoContent)
	}
}

func KeysRefill(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r []database.Key
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		err := coordinator.Database.KeyTable.Refill(r)
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

		go coordinator.Sync()

		return c.NoContent(http.StatusNoContent)
	}
}
