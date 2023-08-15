package prometheus

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type config struct {
	path    string
	content struct {
		Global struct {
			ScrapeInterval string `yaml:"scrape_interval"`
			ExternalLabels struct {
				Monitor string `yaml:"monitor"`
			} `yaml:"external_labels"`
		} `yaml:"global"`
		ScrapeConfigs []*scrapeConfig `yaml:"scrape_configs"`
	}
}

type scrapeConfig struct {
	JobName       string          `yaml:"job_name"`
	StaticConfigs []*staticConfig `yaml:"static_configs"`
}

type staticConfig struct {
	Targets []string `yaml:"targets"`
	Labels  label    `yaml:"labels"`
}

type label struct {
	Service string `yaml:"service"`
}

func (c *config) update(servers map[string]string) error {
	c.content.ScrapeConfigs[0].StaticConfigs = []*staticConfig{}
	for id, s := range servers {
		c.content.ScrapeConfigs[0].StaticConfigs = append(c.content.ScrapeConfigs[0].StaticConfigs, &staticConfig{
			Targets: []string{s}, Labels: label{Service: id},
		})
	}

	content, err := yaml.Marshal(c.content)
	if err != nil {
		return err
	}
	if err = os.WriteFile(c.path, content, 0755); err != nil {
		return errors.New(fmt.Sprintf("cannot save %s, err: %v", c.path, err))
	}
	return nil
}

func newConfig(path string) *config {
	c := &config{}
	c.path = path
	c.content.Global.ScrapeInterval = "5s"
	c.content.Global.ExternalLabels.Monitor = "Shadowsocks"
	c.content.ScrapeConfigs = []*scrapeConfig{
		{
			JobName:       "shadowsocks",
			StaticConfigs: []*staticConfig{},
		},
	}
	return c
}
