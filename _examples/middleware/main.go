package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/actor/middleware"
	"log"
	"reflect"
	"time"
)

type (
	hello struct {
		Who string
	}

	again struct {
	}

	child struct{}
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
	case *hello:
		fmt.Printf("Hello %v\n", msg.Who)
		ctx.Send(ctx.Self(), actor.WrapEnvelop(&again{}))
		ctx.Spawn(actor.PropsFromProducer(createChildActor))
	case *again:
		fmt.Printf("again \n")
	}
}

func senderMiddleware(next actor.SenderFunc) actor.SenderFunc {
	fn := func(c actor.SenderContext, target *actor.PID, envelope *actor.MessageEnvelope) {
		message := envelope.Message
		log.Printf("senderMiddleware %v send %v %+v", c.Self(), reflect.TypeOf(message), message)

		next(c, target, envelope)
	}

	return fn
}

func spawnMiddleware(next actor.SpawnFunc) actor.SpawnFunc {
	fn := func(actorSystem *actor.ActorSystem, id string, props *actor.Props, parentContext actor.SpawnerContext) (*actor.PID, error) {
		pid, err := next(actorSystem, id, props, parentContext)

		log.Printf("spawnMiddleware %v spawn %v", parentContext.Self(), pid)

		return pid, err
	}

	return fn
}

func main() {
	system := actor.NewActorSystem()
	rootContext := actor.NewRootContext(system, nil).WithSpawnMiddleware(spawnMiddleware)
	props := actor.PropsFromFunc(
		receive,
		actor.WithReceiverMiddleware(middleware.ReceiveLogger),
		actor.WithSenderMiddleware(senderMiddleware),
		actor.WithSpawnMiddleware(spawnMiddleware),
	)
	pid := rootContext.Spawn(props)
	rootContext.Send(pid, actor.WrapEnvelop(&hello{Who: "Roger"}))

	time.Sleep(time.Second * 10)
}
