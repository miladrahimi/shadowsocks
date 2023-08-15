package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const AppName = "Shadowsocks"
const AppVersion = "v1.0.0"

// Config is the root configuration.
type Config struct {
	HttpServer struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"http_server"`

	HttpClient struct {
		Timeout int `json:"timeout"`
	} `json:"http_client"`

	Prometheus struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"prometheus"`

	Logger struct {
		Level  string `json:"level"`
		Format string `json:"format"`
	} `json:"logger"`
}

// New creates an instance of the Config.
func New(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot load %s, err: %v", path, err))
	}

	var c Config
	err = json.Unmarshal(content, &c)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot validate %s, err: %v", path, err))
	}

	return &c, err
}
