package main

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/router"
)

type myMessage struct{ i int }

func (m *myMessage) Hash() string {
	return strconv.Itoa(m.i)
}

func main() {
	system := actor.NewActorSystem()
	rootContext := system.Root
	rootContext.Logger().Info("Round robin routing:")
	act := func(ctx actor.Context) {
		envelope := ctx.Envelope()
		switch msg := envelope.Message.(type) {
		case *myMessage:
			ctx.Logger().Info("got message", slog.Any("self", ctx.Self()), slog.Any("message", msg))
		}
	}

	pid := rootContext.Spawn(router.NewRoundRobinPool(5, actor.WithFunc(act)))
	for i := 0; i < 10; i++ {
		rootContext.Send(pid, actor.WrapEnvelope(&myMessage{i}))
	}
	time.Sleep(1 * time.Second)
	rootContext.Stop(pid)
	system.Logger().Info("Random routing:")
	pid = rootContext.Spawn(router.NewRandomPool(5, actor.WithFunc(act)))
	for i := 0; i < 10; i++ {
		rootContext.Send(pid, actor.WrapEnvelope(&myMessage{i}))
	}
	time.Sleep(1 * time.Second)
	rootContext.Stop(pid)
	system.Logger().Info("ConsistentHash routing:")
	pid = rootContext.Spawn(router.NewConsistentHashPool(5, actor.WithFunc(act)))
	for i := 0; i < 10; i++ {
		rootContext.Send(pid, actor.WrapEnvelope(&myMessage{i}))
	}
	time.Sleep(1 * time.Second)
	rootContext.Stop(pid)
	system.Logger().Info("BroadcastPool routing:")
	pid = rootContext.Spawn(router.NewBroadcastPool(5, actor.WithFunc(act)))
	for i := 0; i < 10; i++ {
		rootContext.Send(pid, actor.WrapEnvelope(&myMessage{i}))
	}

	time.Sleep(1 * time.Second)
	rootContext.Stop(pid)
	system.Shutdown()
}
