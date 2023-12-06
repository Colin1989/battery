package agent

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/logger"
	"log/slog"
	"reflect"
)

type pendingWrite struct {
	data []byte
	err  error
}

type Agent struct {
	ctx     actor.Context
	conn    facade.Connector
	session interface{}
	chSend  chan pendingWrite // push message queue
	chDie   chan struct{}     // wait for close
	//packetDecoder codec.PacketDecoder
	//packetEncoder codec.PacketEncoder
	//serializer    serialize.Serializer
}

func NewAgent(conn facade.Connector) actor.Actor {
	return &Agent{
		conn:   conn,
		chSend: make(chan pendingWrite),
		chDie:  make(chan struct{}),
		//packetDecoder: nil,
		//packetEncoder: nil,
		//serializer:    nil,
	}
}

func (a *Agent) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		a.ctx = ctx
		a.run()
	case *actor.Restarting:
		logger.Debug("actor restarting", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		close(a.chDie)
		a.conn.Close()
	//case *actor.ReceiveTimeout:
	//	ctx.Stop(ctx.Self())
	default:
		// TODO Router
		logger.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}

func (a *Agent) send(data []byte) (err error) {
	return nil
}

func (a *Agent) run() {
	go a.write()
	go a.read()
}

func (a *Agent) write() {
	defer func() {
		close(a.chSend)
	}()

	for {
		select {
		case pWrite, ok := <-a.chSend:
			if !ok {
				return
			}
			if _, err := a.conn.Write(pWrite.data); err != nil {
				logger.Error("Failed to write in conn", slog.String("err", err.Error()))
				return
			}
		case <-a.chDie:
			return
		}
	}
}

func (a *Agent) read() {
	defer func() {

	}()

	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			logger.Error("conn receive error",
				slog.String("pid", a.ctx.Self().String()),
				slog.String("err", err.Error()))
			a.ctx.Send(a.ctx.Self(), actor.WrapEnvelop(err))
			//ctx.Poison(ctx.Self())
			return
		}
		a.ctx.Send(a.ctx.Self(), actor.WrapEnvelop(msg))
	}
}
