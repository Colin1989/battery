package battery

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/net/acceptor"
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

func WithTCPAcceptor(addr string, certs ...string) Option {
	return func(app *Application) error {
		producer := actor.PropsFromProducer(
			func() actor.Actor {
				return acceptor.NewTCPAcceptor(addr, certs...)
			})
		tcpPID, err := app.system.Root.SpawnNamed(producer, constant.TCPAcceptor)
		if err != nil {
			return err
		}

		app.actors.Add(tcpPID)

		return nil
	}
}

func WithWSAcceptor(addr string, certs ...string) Option {
	return func(app *Application) error {
		producer := actor.PropsFromProducer(
			func() actor.Actor {
				return acceptor.NewWSAcceptor(addr, certs...)
			})
		wsPID, err := app.system.Root.SpawnNamed(producer, constant.WSAcceptor)
		if err != nil {
			return err
		}

		app.actors.Add(wsPID)

		return nil
	}
}
