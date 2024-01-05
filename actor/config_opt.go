package actor

import "log/slog"

type ConfigOption func(config *Config)

func Configure(options ...ConfigOption) *Config {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithLoggerFactory sets the logger factory to use for the actor system
func WithLoggerFactory(factory func(system *ActorSystem) *slog.Logger) ConfigOption {
	return func(config *Config) {
		config.LoggerFactory = factory
	}
}
