package main

import (
	"fmt"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/wrap"
	"github.com/colin1989/battery/router"
)

type Room struct {
	app            facade.App
	broadcastGroup *actor.PID
}

func NewRoomService(app facade.App) *Room {
	return &Room{app: app}
}

func (r *Room) App() facade.App {
	return r.app
}

func (r *Room) Name() string {
	return "room"
}

func (r *Room) OnStart(ctx actor.Context) {
	r.broadcastGroup = ctx.Spawn(router.NewBroadcastGroup())
}

func (r *Room) OnDestroy(ctx actor.Context) {
}

func (r *Room) PushMembers(sender *actor.PID, ctx actor.Context) {
	request, err := ctx.Request(r.broadcastGroup, router.GetRouteesEnvelope())
	if err != nil {
		return
	}

	routees, ok := request.Message.(*router.Routees)
	if !ok {
		return
	}

	allMembers := make([]string, 0, len(routees.PIDs))
	for _, pid := range routees.PIDs {
		allMembers = append(allMembers, pid.ID)
	}

	ctx.Send(sender, wrap.WrapPushEnvelop("onMembers", &AllMembers{Members: allMembers}))
}

func (r *Room) Join(ctx actor.Context) (*JoinResponse, error) {
	response := &JoinResponse{
		Code:   0,
		Result: "success",
	}
	go r.PushMembers(ctx.Sender(), ctx)
	// ctx.Send(ctx.Sender(), actor.WrapResponseEnvelop(msg.ID, response))

	push := wrap.WrapBroadcast(r.app, "onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", ctx.Sender().String())})
	ctx.Send(r.broadcastGroup, router.BroadcastMessageEnvelope(push))

	ctx.Send(r.broadcastGroup, router.AddRouteeEnvelope(ctx.Sender()))

	//_ = response
	return response, nil
}

func (r *Room) Message(ctx actor.Context, message *UserMessage) {
	push := wrap.WrapBroadcast(r.app, "onMessage", message)
	ctx.Send(r.broadcastGroup, router.BroadcastMessageEnvelope(push))
}
