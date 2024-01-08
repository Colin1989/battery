package battery

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/gate"
)

// Option is a function on the options for a connection.
type Option func(app *Application) error

func WithDebug() Option {
	return func(app *Application) error {
		app.debug = true
		return nil
	}
}

func WithClusterMode() Option {
	return func(app *Application) error {
		app.serverMode = Cluster
		return nil
	}
}

func WithGate(acceptors []facade.Acceptors) Option {
	return func(app *Application) error {
		producer := actor.PropsFromProducer(
			func() actor.Actor {
				return gate.NewGate(acceptors, app)
			})
		pid, err := app.system.Root.SpawnNamed(producer, constant.Gate)
		if err != nil {
			return err
		}
		app.actors.Add(pid)

		return nil
	}
}
