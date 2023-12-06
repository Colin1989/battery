package agent

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/logger"
	"github.com/colin1989/battery/message"
	"log/slog"
	"reflect"
)

//goland:noinspection GoNameStartsWithPackageName
type AgentManager struct {
}

func NewAgentManager() actor.Actor {
	return &AgentManager{}
}

func (am *AgentManager) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Restarting:
		logger.Debug("actor restarting", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	case *message.NewAgent:
		props := actor.PropsFromProducer(func() actor.Actor {
			return NewAgent(msg.Conn)
		})
		pid := ctx.SpawnPrefix(props, constant.AgentPrefix)
		_ = pid
	default:
		logger.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}
