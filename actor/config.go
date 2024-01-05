package actor

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type Config struct {
	LoggerFactory func(system *ActorSystem) *slog.Logger
}

func defaultConfig() *Config {
	return &Config{
		LoggerFactory: func(system *ActorSystem) *slog.Logger {
			w := os.Stderr

			// create a new logger
			return slog.New(tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: time.Kitchen,
			})).With("lib", "Proto.Actor").
				With("system", system.ID)
		}}
}
