package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"strings"
)

func Authorize(d *database.Database) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			if !authorizeToken(d.SettingTable.ApiToken, context) {
				return echo.ErrUnauthorized
			}
			return next(context)
		}
	}
}

func authorizeToken(token string, context echo.Context) bool {
	authHeader := context.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[len("Bearer "):] == token
	}
	return false
}
