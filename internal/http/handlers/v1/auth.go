package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"net/http"
	"time"
)

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SignIn(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			time.Sleep(time.Second)
		}()

		var r SignInRequest
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		if r.Username == "admin" && r.Password == coordinator.Database.SettingTable.AdminPassword {
			return c.JSON(http.StatusOK, map[string]string{
				"token": coordinator.Database.SettingTable.ApiToken,
			})
		}

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "Unauthorized.",
		})
	}
}
