package main

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/logger"
)

type Room struct {
	users actor.PIDSet
}

func (r *Room) Name() string {
	return "room"
}

func (r *Room) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("room started", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("room stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("room stopped", slog.String("pid", ctx.Self().String()))
	//case *message.Message:
	//	r.ProcessMessage(ctx, msg)
	case *actor.DeadLetterResponse:
		r.users.Remove(msg.Target)
		logger.Debug("room DeadLetterResponse", slog.String("pid", ctx.Self().String()))
	default:
		logger.Warn("room unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}

//func (r *Room) OnStart(as facade.ActorService) {
//	//as.RegisterHandler(&Join{}, r.Join)
//}

func (r *Room) AllMembers() []string {
	allMembers := make([]string, 0, r.users.Len())
	r.users.ForEach(func(_ int, pid *actor.PID) {
		allMembers = append(allMembers, pid.ID)
	})
	return allMembers
}

func (r *Room) Join(ctx actor.Context) (*JoinResponse, error) {
	response := &JoinResponse{
		Code:   0,
		Result: "success",
	}
	// ctx.Send(ctx.Sender(), actor.WrapResponseEnvelop(msg.ID, response))

	ctx.Send(ctx.Sender(), actor.WrapPushEnvelop("onMembers", &AllMembers{Members: r.AllMembers()}))

	r.users.ForEach(func(_ int, pid *actor.PID) {
		ctx.Send(pid, actor.WrapPushEnvelop("onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", ctx.Sender().String())}))
	})

	r.users.Add(ctx.Sender())

	//_ = response
	return response, nil
}

func (r *Room) Message(ctx actor.Context, message *UserMessage) {
	r.users.ForEach(func(_ int, pid *actor.PID) {
		ctx.Send(pid, actor.WrapPushEnvelop("onMessage", message))
	})
}
