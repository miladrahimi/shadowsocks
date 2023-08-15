package prometheus

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Prometheus struct {
	http   *http.Client
	logger *zap.Logger
	config *config
	host   string
	port   int
}

func (p *Prometheus) Metrics() (*Stats, error) {
	query := `sum(increase(shadowsocks_data_bytes{dir=~"c<p|c>p"}[30d]))%20by%20(access_key,proto,dir,service)`
	url := fmt.Sprintf("http://%s:%d/api/v1/query?query=%s", p.host, p.port, query)
	response, err := p.http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("unknown healthz status %s", response.Status))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var s Stats
	return &s, json.Unmarshal(body, &s)
}

func (p *Prometheus) Reload() {
	url := fmt.Sprintf("http://%s:%d/-/reload", p.host, p.port)
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		p.logger.Error("cannot create prometheus reload request", zap.String("url", url), zap.Error(err))
		return
	}

	response, err := p.http.Do(request)
	if err != nil {
		p.logger.Error("cannot request prometheus to reload", zap.String("url", url), zap.Error(err))
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		p.logger.Debug("prometheus reloaded successfully")
	} else {
		p.logger.Error("prometheus reload failed", zap.Error(err))
	}
}

func (p *Prometheus) Update(servers map[string]string) error {
	return p.config.update(servers)
}

func New(l *zap.Logger, hc *http.Client, cp, host string, port int) *Prometheus {
	return &Prometheus{
		config: newConfig(cp),
		logger: l,
		http:   hc,
		host:   host,
		port:   port,
	}
}
