package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
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
		fmt.Println("actor started")
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *actor.Stopping:
		fmt.Println("actor stopping")
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
		fmt.Printf("receive deadleeter : %v", dlEvent)
	})
	defer func() {
		system.EventStream.Unsubscribe(eventSub)
	}()
	system.Root.Send(pid, actor.WrapEnvelop(&hello{Say: "Poison"}))

	system.Shutdown()
}
