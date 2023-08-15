package handlers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"net/http"
)

func Metrics(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		url := fmt.Sprintf("http://127.0.0.1:%d%s", coordinator.MetricsPort, c.Request().RequestURI)
		r, err := http.Get(url)
		if err != nil {
			return err
		}
		return c.Stream(r.StatusCode, r.Header.Get("Content-Type"), r.Body)
	}
}
