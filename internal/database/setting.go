package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/miladrahimi/shadowsocks/pkg/utils"
	"os"
	"path/filepath"
)

const SettingPath = "storage/database/settings.json"

type SettingTable struct {
	AdminPassword      string  `json:"admin_password" validate:"required,min=8,max=32"`
	ApiToken           string  `json:"api_token" validate:"required,min=16,max=128"`
	ShadowsocksEnabled bool    `json:"shadowsocks_enabled"`
	ShadowsocksHost    string  `json:"shadowsocks_host" validate:"required,max=128"`
	ShadowsocksPort    int     `json:"shadowsocks_port" validate:"required,min=1,max=65536"`
	ExternalHttps      string  `json:"external_https"`
	ExternalHttp       string  `json:"external_http" validate:"required"`
	TrafficRatio       float64 `json:"traffic_ratio" validate:"required,min=1"`
}

func (st *SettingTable) Load() error {
	content, err := os.ReadFile(SettingPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !utils.DirectoryExist(filepath.Dir(SettingPath)) {
				return errors.New(fmt.Sprintf("directory %s not found", filepath.Base(SettingPath)))
			}
			return st.Save()
		}
		return errors.New(fmt.Sprintf("cannot load %s, err: %v", SettingPath, err))
	}

	err = json.Unmarshal(content, st)
	if err != nil {
		return err
	}

	if err = validator.New().Struct(st); err != nil {
		return errors.New(fmt.Sprintf("cannot validate %s, err: %v", SettingPath, err))
	}

	return nil
}

func (st *SettingTable) Save() error {
	if err := validator.New().Struct(st); err != nil {
		return DataError(err.Error())
	}

	content, err := json.Marshal(st)
	if err != nil {
		return err
	}

	if err = os.WriteFile(SettingPath, content, 0755); err != nil {
		return errors.New(fmt.Sprintf("cannot save %s, err: %v", ServerPath, err))
	}

	return st.Load()
}
