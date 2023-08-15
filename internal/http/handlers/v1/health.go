package v1

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func Health() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}
}
