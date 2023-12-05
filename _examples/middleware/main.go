package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/actor/middleware"
	"log"
	"math/rand"
	"reflect"
)

type (
	hello struct {
		Who string
	}

	again struct {
	}

	child struct{}

	middleWare1 struct {
		RandNum int
	}
)

func (c *child) Receive(ctx actor.Context) {
	fmt.Printf("Receive [%v] \n", ctx.Envelope())
}

func createChildActor() actor.Actor {
	return &child{}
}

func receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started")
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *actor.Stopping:
		fmt.Println("actor stopping")
	case *hello:
		fmt.Printf("Hello %v\n", msg.Who)
		ctx.Send(ctx.Self(), actor.WrapEnvelop(&again{}))
		ctx.Spawn(actor.PropsFromProducer(createChildActor))
	case *again:
		fmt.Printf("again \n")
	}
}

func (mw *middleWare1) senderMiddleware(next actor.SenderFunc) actor.SenderFunc {
	fn := func(c actor.SenderContext, target *actor.PID, envelope *actor.MessageEnvelope) {
		message := envelope.Message
		log.Printf("senderMiddleware %v send %v %+v", c.Self(), reflect.TypeOf(message), message)

		next(c, target, envelope)
	}

	return fn
}

func (mw *middleWare1) spawnMiddleware(next actor.SpawnFunc) actor.SpawnFunc {
	fn := func(actorSystem *actor.ActorSystem, id string, props *actor.Props, parentContext actor.SpawnerContext) (*actor.PID, error) {
		pid, err := next(actorSystem, id, props, parentContext)

		log.Printf("spawnMiddleware %v spawn %v", parentContext.Self(), pid)

		return pid, err
	}

	return fn
}

func main() {
	system := actor.NewActorSystem()
	mw := new(middleWare1)
	mw.RandNum = rand.Int()
	fmt.Printf("RandNum : [%v] \n", mw.RandNum)
	rootContext := actor.NewRootContext(system, nil).WithSpawnMiddleware(mw.spawnMiddleware)
	props := actor.PropsFromFunc(
		receive,
		actor.WithReceiverMiddleware(middleware.ReceiveLogger),
		actor.WithSenderMiddleware(mw.senderMiddleware),
		actor.WithSpawnMiddleware(mw.spawnMiddleware),
	)
	pid := rootContext.Spawn(props)
	rootContext.Send(pid, actor.WrapEnvelop(&hello{Who: "Roger"}))
	rootContext.Send(pid, actor.WrapEnvelop(&hello{Who: "Roger"}))
	rootContext.Poison(pid)

	system.Shutdown()
}
