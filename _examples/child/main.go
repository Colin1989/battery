package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"sync"
	"time"
)

type (
	foo                struct{}
	child              struct{}
	MessageCreateChild struct{}
)

func createChildActor() actor.Actor {
	return &child{}
}

func (c *child) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started child")
	case *actor.Stopped:
		fmt.Println("actor stopped child")
		ctx.Poison(parent)
		wg.Done()
	default:
		fmt.Printf("unsupported msg : %+v \n", msg)
	}
}

func (f *foo) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started")
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *MessageCreateChild:
		childPID = ctx.Spawn(actor.PropsFromProducer(createChildActor))
	default:
		fmt.Printf("unsupported msg : %+v \n", msg)
	}
}

var (
	parent   *actor.PID
	childPID *actor.PID
	wg       sync.WaitGroup
)

func main() {
	wg.Add(1)
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &foo{}
	})

	parent = system.Root.Spawn(props)
	system.Root.Send(parent, actor.WrapEnvelop(&MessageCreateChild{}))

	time.Sleep(time.Second * 1)
	if childPID != nil {
		system.Root.Poison(childPID)
	}
	time.Sleep(time.Second * 1)
	wg.Wait()
}
