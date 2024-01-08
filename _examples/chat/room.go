package main

import (
	"fmt"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/router"
)

type Room struct {
	broadcastGroup *actor.PID
}

func NewRoomService() *Room {

	return &Room{}
}

func (r *Room) Name() string {
	return "room"
}

//func (r *Room) Receive(ctx actor.Context) {
//	envelope := ctx.Envelope()
//	switch msg := envelope.Message.(type) {
//	//	case *actor.Started:
//	//		blog.Debug("room started", slog.String("pid", ctx.Self().String()))
//	//	case *actor.Stopping:
//	//		blog.Debug("room stopping", slog.String("pid", ctx.Self().String()))
//	//	case *actor.Stopped:
//	//		blog.Debug("room stopped", slog.String("pid", ctx.Self().String()))
//	//	//case *message.Message:
//	//	//	r.ProcessMessage(ctx, msg)
//	//	case *actor.Terminated:
//	//		r.users.Remove(msg.Who)
//	//		blog.Debug("room DeadLetterResponse", slog.String("pid", ctx.Self().String()))
//	//	case *actor.DeadLetterResponse:
//	//		r.users.Remove(msg.Target)
//	//		blog.Debug("room DeadLetterResponse", slog.String("pid", ctx.Self().String()))
//	default:
//		blog.Warn("room unsupported type",
//			blog.TypeAttr(msg),
//			slog.Any("msg", msg))
//	}
//}

func (r *Room) OnStart(ctx actor.Context) {
	//as.RegisterHandler(&Join{}, r.Join)
	r.broadcastGroup = ctx.Spawn(router.NewBroadcastGroup())
}

func (r *Room) OnDestroy(ctx actor.Context) {
	//ctx.Stop(r.broadcastGroup)
}

func (r *Room) AllMembers(ctx actor.Context) []string {
	request, err := ctx.Request(r.broadcastGroup, router.GetRouteesEnvelope())
	if err != nil {
		return nil
	}

	routees, ok := request.Message.(*router.Routees)
	if !ok {
		return nil
	}

	allMembers := make([]string, 0, len(routees.PIDs))
	for _, pid := range routees.PIDs {
		allMembers = append(allMembers, pid.ID)
	}
	return allMembers
}

func (r *Room) Join(ctx actor.Context) (*JoinResponse, error) {
	response := &JoinResponse{
		Code:   0,
		Result: "success",
	}
	// ctx.Send(ctx.Sender(), actor.WrapResponseEnvelop(msg.ID, response))
	ctx.Send(ctx.Sender(), actor.WrapPushEnvelop("onMembers", &AllMembers{Members: r.AllMembers(ctx)}))

	push := actor.WrapPushEnvelop("onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", ctx.Sender().String())})
	ctx.Send(r.broadcastGroup, push)

	ctx.Send(r.broadcastGroup, router.AddRouteeEnvelope(ctx.Sender()))

	//_ = response
	return response, nil
}

func (r *Room) Message(ctx actor.Context, message *UserMessage) {
	push := actor.WrapPushEnvelop("onMessage", message)
	ctx.Send(r.broadcastGroup, push)
}
