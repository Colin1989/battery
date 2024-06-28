package config

import "time"

type EtcdConfig struct {
	BaseKey      string        //`json:"basekey"`
	Endpoints    []string      //`json:"endpoints"`
	Username     string        //`json:"username"`
	Password     string        //`json:"password"`
	DialTimeout  time.Duration //`json:"dialtimeout"`
	Prefix       string        //`json:"prefix"`
	HeartbeatTTL time.Duration //`json:"ttl"`
}

func DefaultEtcdConfig() EtcdConfig {
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

type NatsConfig struct {
	Address        string
	MaxReconnects  int
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

func DefaultNatsConfig() NatsConfig {
	return NatsConfig{
		Address:        "nats://localhost:4222",
		MaxReconnects:  15,
		ConnectTimeout: 5 * time.Second,
		RequestTimeout: 5 * time.Second,
	}
}
