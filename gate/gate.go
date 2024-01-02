package gate

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/agent"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/logger"
	"github.com/colin1989/battery/net/acceptor"
	"log/slog"
	"reflect"
)

//goland:noinspection GoNameStartsWithPackageName
type Gate struct {
	pid       *actor.PID
	acceptors []facade.Acceptors

	app facade.App
}

func NewGate(acceptors []facade.Acceptors, app facade.App) *Gate {
	ga := &Gate{
		acceptors: acceptors,
		app:       app,
	}
	return ga
}

func (gs *Gate) addTCPAcceptor(ctx actor.Context, addr string, certs ...string) error {
	producer := actor.PropsFromProducer(
		func() actor.Actor {
			tcpAcc := acceptor.NewTCPAcceptor(addr, certs...)
			return tcpAcc
		})
	_, err := ctx.SpawnNamed(producer, constant.TCPAcceptor)
	return err
}

func (gs *Gate) addWSAcceptor(ctx actor.Context, addr string, certs ...string) error {
	producer := actor.PropsFromProducer(
		func() actor.Actor {
			wsAcc := acceptor.NewWSAcceptor(addr, certs...)
			return wsAcc
		})
	_, err := ctx.SpawnNamed(producer, constant.WSAcceptor)
	return err
}

func (gs *Gate) OnStarted(ctx actor.Context) {
	for _, acc := range gs.acceptors {
		var err error
		switch acc.AcceptorType {
		case constant.AcceptorTypeTCP:
			err = gs.addTCPAcceptor(ctx, acc.Addr, acc.Certs[0], acc.Certs[1])
		case constant.AcceptorTypeWS:
			err = gs.addWSAcceptor(ctx, acc.Addr, acc.Certs[0], acc.Certs[1])
		}
		if err != nil {
			logger.Fatal("new acceptor error", slog.Any("acceptor", acc), logger.ErrAttr(err))
		}
	}
}

func (gs *Gate) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		gs.OnStarted(ctx)
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Restarting:
		logger.Debug("actor restarting", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	case facade.Connector:
		conn := msg
		props := actor.PropsFromProducer(func() actor.Actor {
			return agent.NewAgent(conn, gs.app)
		})
		pid := ctx.SpawnPrefix(props, constant.AgentPrefix)
		_ = pid
	default:
		logger.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}
