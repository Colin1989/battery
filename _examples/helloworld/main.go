package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/logger"
	"log/slog"
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
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	case *hello:
		logger.Info("Hello", slog.String("say", msg.Say))
	}
}

func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &helloActor{}
	})

	pid := system.Root.Spawn(props)
	system.Root.Send(pid, actor.WrapEnvelop(&hello{Say: "World"}))

	for i := 0; i < 10000; i++ {
		j := i
		go func() {
			system.Root.Send(pid, actor.WrapEnvelop(&hello{Say: fmt.Sprintf("ID:%v", j)}))
		}()
	}
	system.Root.Poison(pid)

	eventSub := system.EventStream.Subscribe(func(msg actor.EventMessage) {
		dlEvent, ok := msg.(*actor.DeadLetterEvent)
		if !ok {
			return
		}
		logger.Info("receive dead letter", slog.Any("event", dlEvent))
	})
	defer func() {
		system.EventStream.Unsubscribe(eventSub)
	}()
	system.Root.Send(pid, actor.WrapEnvelop(&hello{Say: "Poison"}))

	system.Shutdown()
}
