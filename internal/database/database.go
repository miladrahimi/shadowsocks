package database

type Database struct {
	SettingTable *SettingTable
	KeyTable     *KeyTable
	ServerTable  *ServerTable
}

func New() (*Database, error) {
	db := &Database{
		SettingTable: &SettingTable{
			AdminPassword:      "password",
			ApiToken:           "api-token-secret",
			ShadowsocksHost:    "127.0.0.1",
			ShadowsocksPort:    1,
			ShadowsocksEnabled: true,
			ExternalHttps:      "",
			ExternalHttp:       "http://localhost",
			TrafficRatio:       1,
		},
		KeyTable: &KeyTable{
			Keys:   []*Key{},
			NextId: 1,
		},
		ServerTable: &ServerTable{
			Servers: []*Server{},
			NextId:  1,
		},
	}

	if err := db.SettingTable.Load(); err != nil {
		return nil, err
	}
	if err := db.KeyTable.Load(); err != nil {
		return nil, err
	}
	if err := db.ServerTable.Load(); err != nil {
		return nil, err
	}

	return db, nil
}
