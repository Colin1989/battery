package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"time"
)

type (
	parent             struct{}
	child              struct{}
	MessageCreateChild struct{}
)

func createChildActor() actor.Actor {
	return &child{}
}

func (c *child) Receive(ctx actor.Context) {
	childCtx = ctx
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started child")
	case *actor.Stopped:
		fmt.Println("actor stopped child")
	default:
		fmt.Printf("unsupported type %T msg : %+v \n", msg, msg)
	}
}

func (f *parent) Receive(ctx actor.Context) {
	parentCtx = ctx
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started")
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *actor.Stopping:
		fmt.Println("actor stopping")
	case *MessageCreateChild:
		childPID = ctx.Spawn(actor.PropsFromProducer(createChildActor))
	default:
		fmt.Printf("unsupported type %T msg : %+v \n", msg, msg)
	}
}

var (
	parentPID *actor.PID
	parentCtx actor.Context
	childPID  *actor.PID
	childCtx  actor.Context
)

func main() {
	system := actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &parent{}
	})

	parentPID = system.Root.Spawn(props)
	system.Root.Send(parentPID, actor.WrapEnvelop(&MessageCreateChild{}))

	time.Sleep(time.Second * 1)
	if len(parentCtx.Children()) != 1 {
		panic("children count is not 1")
	}
	if childCtx.Parent() != parentPID {
		panic("the child parent PID is not equal parentPID")
	}
	system.Root.Poison(childPID)
	time.Sleep(time.Second * 1)
	if len(parentCtx.Children()) != 0 {
		panic("children count is not 0")
	}
	system.Root.Poison(parentPID)
	system.Shutdown()
}
