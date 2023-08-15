package coordinator

import (
	"github.com/miladrahimi/shadowsocks/pkg/shadowsocks"
	"go.uber.org/zap"
	"time"
)

func (c *Coordinator) syncKeys(reconfigure bool) {
	c.Logger.Debug("syncing keys with the shadowsocks server...")

	keys := make([]shadowsocks.Key, 0, len(c.Database.KeyTable.Keys))
	for _, k := range c.Database.KeyTable.Keys {
		if !k.Enabled {
			continue
		}
		keys = append(keys, shadowsocks.Key{
			Id:     k.Id,
			Secret: k.Secret,
			Cipher: k.Cipher,
			Port:   c.Database.SettingTable.ShadowsocksPort,
		})
	}

	if err := c.Shadowsocks.Update(keys); err != nil {
		c.Logger.Fatal("cannot sync keys with the shadowsocks server", zap.Error(err))
	}

	if reconfigure {
		c.Shadowsocks.Reconfigure()
	}

	c.SyncedAt = time.Now().Unix()
}
