package main

import (
	"fmt"
	"log/slog"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/blog"
)

type (
	hello struct {
		Say string
	}
	helloActor struct{}
)

func (h *helloActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		blog.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		blog.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		blog.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	case *hello:
		blog.Info("Hello", slog.String("say", msg.Say))
	}
}

func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &helloActor{}
	})

	pid := system.Root.Spawn(props)
	system.Root.Send(pid, actor.WrapEnvelope(&hello{Say: "World"}))

	for i := 0; i < 10000; i++ {
		j := i
		go func() {
			system.Root.Send(pid, actor.WrapEnvelope(&hello{Say: fmt.Sprintf("ID:%v", j)}))
		}()
	}
	system.Root.Poison(pid)

	eventSub := system.EventStream.Subscribe(func(msg actor.EventMessage) {
		dlEvent, ok := msg.(*actor.DeadLetterEvent)
		if !ok {
			return
		}
		blog.Info("receive dead letter", slog.Any("event", dlEvent))
	})
	defer func() {
		system.EventStream.Unsubscribe(eventSub)
	}()
	system.Root.Send(pid, actor.WrapEnvelope(&hello{Say: "Poison"}))

	system.Shutdown()
}
