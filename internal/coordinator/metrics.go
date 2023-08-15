package coordinator

import (
	"go.uber.org/zap"
	"strconv"
)

type ServerMetric struct {
	Id      string `json:"id"`
	DownTcp int64  `json:"down_tcp"`
	UpTcp   int64  `json:"up_tcp"`
	DownUdp int64  `json:"down_udp"`
	UpUdp   int64  `json:"up_udp"`
	Total   int64  `json:"total"`
}

type KeyMetric struct {
	Id      string `json:"id"`
	DownTcp int64  `json:"down_tcp"`
	UpTcp   int64  `json:"up_tcp"`
	DownUdp int64  `json:"down_udp"`
	UpUdp   int64  `json:"up_udp"`
	Total   int64  `json:"total"`
}

func (c *Coordinator) syncMetrics() {
	c.Logger.Debug("syncing metrics...")

	metrics, err := c.Prometheus.Metrics()
	if err != nil {
		c.Logger.Error("prometheus query failed", zap.Error(err))
		return
	}

	sms := map[string]*ServerMetric{}
	kms := map[string]*KeyMetric{}

	for _, r := range metrics.Data.Result {
		f, err := strconv.ParseFloat(r.Value[1].(string), 64)
		if err != nil {
			c.Logger.Error("cannot parse prometheus metric", zap.Error(err))
			continue
		}
		v := int64(f)

		if _, found := sms[r.Metric.Service]; !found {
			sms[r.Metric.Service] = &ServerMetric{Id: r.Metric.Service}
		}

		if _, found := kms[r.Metric.AccessKey]; !found {
			kms[r.Metric.AccessKey] = &KeyMetric{Id: r.Metric.AccessKey}
		}

		if r.Metric.Dir == "c<p" && r.Metric.Proto == "tcp" {
			sms[r.Metric.Service].DownTcp += v
			kms[r.Metric.AccessKey].DownTcp += v
		} else if r.Metric.Dir == "c<p" && r.Metric.Proto == "udp" {
			sms[r.Metric.Service].DownUdp += v
			kms[r.Metric.AccessKey].DownUdp += v
		} else if r.Metric.Dir == "c>p" && r.Metric.Proto == "tcp" {
			sms[r.Metric.Service].UpTcp += v
			kms[r.Metric.AccessKey].UpTcp += v
		} else if r.Metric.Dir == "c>p" && r.Metric.Proto == "udp" {
			sms[r.Metric.Service].UpUdp += v
			kms[r.Metric.AccessKey].UpUdp += v
		}

		sms[r.Metric.Service].Total += v
		kms[r.Metric.AccessKey].Total += v
	}

	c.ServerMetrics = sms
	c.KeyMetrics = kms

	go c.checkQuotas()
}

func (c *Coordinator) checkQuotas() {
	dirty := false
	for _, k := range c.Database.KeyTable.Keys {
		if !k.Enabled {
			continue
		}

		if m, found := c.KeyMetrics[k.Id]; found {
			if k.Quota != 0 && m.Total/1000000 > k.Quota {
				k.Enabled = false
				if _, err := c.Database.KeyTable.Update(*k); err != nil {
					c.Logger.Error("cannot update the key", zap.Error(err))
				} else {
					dirty = true
				}
			}
		}
	}

	if dirty {
		go c.Sync()
	}
}
