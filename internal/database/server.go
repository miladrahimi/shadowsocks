package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/miladrahimi/shadowsocks/pkg/utils"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"time"
)

const ServerPath = "storage/database/servers.json"

const (
	ServerStatusActive       = "active"
	ServerStatusProcessing   = "processing"
	ServerStatusUnauthorized = "unauthorized"
	ServerStatusUnavailable  = "unavailable"
)

type Server struct {
	Id                 string `json:"id" validate:"required"`
	HttpHost           string `json:"http_host" validate:"required"`
	HttpPort           int    `json:"http_port" validate:"required,min=1,max=65536"`
	ShadowsocksEnabled bool   `json:"shadowsocks_enabled"`
	ShadowsocksHost    string `json:"shadowsocks_host"`
	ShadowsocksPort    int    `json:"shadowsocks_port" validate:"min=1,max=65536"`
	ApiToken           string `json:"api_token"`
	Status             string `json:"status"`
	SyncedAt           int64  `json:"synced_at" validate:"min=0"`
}

type ServerTable struct {
	Servers   []*Server `json:"keys" validate:"required"`
	NextId    int64     `json:"next_id" validate:"required,min=1"`
	UpdatedAt int64     `json:"updated_at" validate:"min=0"`
}

func (st *ServerTable) Load() error {
	content, err := os.ReadFile(ServerPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !utils.DirectoryExist(filepath.Dir(ServerPath)) {
				return errors.New(fmt.Sprintf("directory %s not found", filepath.Base(ServerPath)))
			}
			return st.Save()
		}
		return errors.New(fmt.Sprintf("cannot load %s, err: %v", ServerPath, err))
	}

	err = json.Unmarshal(content, st)
	if err != nil {
		return err
	}

	if err = validator.New().Struct(st); err != nil {
		return errors.New(fmt.Sprintf("cannot validate %s, err: %v", ServerPath, err))
	}

	return nil
}

func (st *ServerTable) Save() (err error) {
	if err = validator.New().Struct(st); err != nil {
		return DataError(err.Error())
	}
	for _, s := range st.Servers {
		if err = validator.New().Struct(s); err != nil {
			return DataError(err.Error())
		}
	}

	st.UpdatedAt = time.Now().Unix()
	content, err := json.Marshal(st)
	if err != nil {
		return err
	}

	if err = os.WriteFile(ServerPath, content, 0755); err != nil {
		return errors.New(fmt.Sprintf("cannot save %s, err: %v", ServerPath, err))
	}

	return st.Load()
}

func (st *ServerTable) Store(server Server) (*Server, error) {
	server.Id = fmt.Sprintf("s-%d", st.NextId)
	server.Status = ServerStatusProcessing
	server.ShadowsocksEnabled = false
	server.ShadowsocksHost = ""
	server.ShadowsocksPort = 1

	st.NextId++
	st.Servers = append(st.Servers, &server)

	return &server, st.Save()
}

func (st *ServerTable) Update(server Server) (*Server, error) {
	for i, s := range st.Servers {
		if s.Id == server.Id {
			st.Servers[i].HttpHost = server.HttpHost
			st.Servers[i].HttpPort = server.HttpPort
			st.Servers[i].ShadowsocksEnabled = server.ShadowsocksEnabled
			st.Servers[i].ShadowsocksHost = server.ShadowsocksHost
			st.Servers[i].ShadowsocksPort = server.ShadowsocksPort
			st.Servers[i].ApiToken = server.ApiToken
			st.Servers[i].Status = server.Status
			return st.Servers[i], st.Save()
		}
	}
	return nil, nil
}

func (st *ServerTable) Find(Id string) *Server {
	for i, s := range st.Servers {
		if s.Id == Id {
			return st.Servers[i]
		}
	}
	return nil
}

func (st *ServerTable) Delete(id string) error {
	for i, s := range st.Servers {
		if s.Id == id {
			st.Servers = slices.Delete(st.Servers, i, i+1)
			return st.Save()
		}
	}
	return nil
}
