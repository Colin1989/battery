package actor

type ConfigOption func(config *Config)

func Configure(options ...ConfigOption) *Config {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}
