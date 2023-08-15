package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/shadowsocks/internal/coordinator"
	"github.com/miladrahimi/shadowsocks/internal/database"
	"net/http"
)

type ServerResponse struct {
	database.Server
	Id   string `json:"id"`
	Used int64  `json:"used"`
}

type ServersStoreRequest struct {
	HttpHost string `json:"http_host"`
	HttpPort int    `json:"http_port"`
	ApiToken string `json:"api_token"`
}

type ServersUpdateRequest struct {
	ServersStoreRequest
	Id string `json:"id"`
}

func ServersIndex(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		servers := make([]ServerResponse, 0, len(coordinator.Database.ServerTable.Servers)+1)

		server := ServerResponse{Server: *coordinator.CurrentServer(), Id: "s-0"}
		if m, found := coordinator.ServerMetrics["s-0"]; found {
			server.Used = m.Total / 1000000
		}
		servers = append(servers, server)

		for _, s := range coordinator.Database.ServerTable.Servers {
			server = ServerResponse{Server: *s, Id: s.Id}
			if m, found := coordinator.ServerMetrics[s.Id]; found {
				server.Used = m.Total / 1000000
			}
			servers = append(servers, server)
		}

		return c.JSON(http.StatusOK, servers)
	}
}

func ServersStore(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r ServersStoreRequest
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		server, err := coordinator.Database.ServerTable.Store(database.Server{
			HttpHost: r.HttpHost,
			HttpPort: r.HttpPort,
			ApiToken: r.ApiToken,
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

		sr := ServerResponse{Server: *server, Id: server.Id}

		return c.JSON(http.StatusCreated, sr)
	}
}

func ServersUpdate(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r ServersUpdateRequest
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		server := coordinator.Database.ServerTable.Find(r.Id)
		if server == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "The server not found.",
			})
		}

		server, err := coordinator.Database.ServerTable.Update(database.Server{
			Id:                 r.Id,
			HttpHost:           r.HttpHost,
			HttpPort:           r.HttpPort,
			ApiToken:           r.ApiToken,
			ShadowsocksEnabled: server.ShadowsocksEnabled,
			ShadowsocksHost:    server.ShadowsocksHost,
			ShadowsocksPort:    server.ShadowsocksPort,
			Status:             server.Status,
			SyncedAt:           server.SyncedAt,
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
		if server == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Server not found.",
			})
		}

		go coordinator.Sync()

		sr := ServerResponse{
			Server: *server,
			Id:     server.Id,
		}
		if m, found := coordinator.ServerMetrics[sr.Id]; found {
			sr.Used = m.Total / 1000000
		}

		return c.JSON(http.StatusOK, sr)
	}
}

func ServersDelete(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		err := coordinator.Database.ServerTable.Delete(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "Cannot update the database.",
			})
		}

		go coordinator.Sync()

		return c.NoContent(http.StatusNoContent)
	}
}
