package config

import "time"

type EtcdConfig struct {
	BaseKey      string        `mapstructure:"basekey"`
	Endpoints    []string      `mapstructure:"endpoints"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	DialTimeout  time.Duration `mapstructure:"dialtimeout"`
	Prefix       string        `mapstructure:"prefix"`
	HeartbeatTTL time.Duration `mapstructure:"ttl"`
}

func NewDefaultEtcdConfig() EtcdConfig {
	return EtcdConfig{
		BaseKey:     "battery",
		Endpoints:   []string{"localhost:2379"},
		Username:    "",
		Password:    "",
		DialTimeout: 5 * time.Second,
		Prefix:      "battery/",
		//HeartbeatTTL: 60 * time.Second,
	}
}
