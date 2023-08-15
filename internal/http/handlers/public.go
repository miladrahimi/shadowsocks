package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"net/http"
	"os"
)

func Public(_ *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		content, err := os.ReadFile("web/public.html")
		if err != nil {
			return err
		}

		return c.HTML(http.StatusOK, string(content))
	}
}
