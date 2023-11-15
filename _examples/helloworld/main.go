package main

import (
	"fmt"
	"github.com/colin1989/battery/actor/actor"
	"time"
)

type (
	hello struct {
		Say string
	}
	helloActor struct{}
)

func (h *helloActor) Receive(ctx actor.Context) {
	envelope := ctx.Message()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started")
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *hello:
		fmt.Printf("Hello %v\n", msg.Say)
	}
}

func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &helloActor{}
	})

	pid := system.Root.Spawn(props)
	system.Root.Send(pid, actor.WrapEnvelop(&hello{Say: "World"}))
	system.Root.Poison(pid)

	eventSub := system.EventStream.Subscribe(func(msg actor.EventMessage) {
		dlEvent, ok := msg.(*actor.DeadLetterEvent)
		if !ok {
			return
		}
		fmt.Printf("receive deadleeter : %v", dlEvent)
	})
	defer func() {
		system.EventStream.Unsubscribe(eventSub)
	}()
	system.Root.Send(pid, actor.WrapEnvelop(&hello{Say: "Poison"}))

	time.Sleep(time.Second * 5)
}
