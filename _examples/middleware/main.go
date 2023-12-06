package main

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/actor/middleware"
	"github.com/colin1989/battery/logger"
	"log"
	"log/slog"
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
	logger.Info("Receive", slog.Any("msg", ctx.Envelope()))
}

func createChildActor() actor.Actor {
	return &child{}
}

func receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	case *hello:
		logger.Info("Hello", slog.String("say", msg.Who))
		ctx.Send(ctx.Self(), actor.WrapEnvelop(&again{}))
		ctx.Spawn(actor.PropsFromProducer(createChildActor))
	case *again:
		logger.Info("again")
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

		logger.Info("spawnMiddleware",
			slog.String("parent", parentContext.Self().String()),
			slog.String("child", pid.String()))

		return pid, err
	}

	return fn
}

func main() {
	system := actor.NewActorSystem()
	mw := new(middleWare1)
	mw.RandNum = rand.Int()
	logger.Info("RandNum", slog.Int("num", mw.RandNum))
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
