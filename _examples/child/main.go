package main

import (
	"log/slog"
	"reflect"
	"time"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/blog"
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
		blog.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		blog.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		blog.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	default:
		blog.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}

func (f *parent) Receive(ctx actor.Context) {
	parentCtx = ctx
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		blog.Debug("actor started", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		blog.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		blog.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
	case *MessageCreateChild:
		childPID = ctx.Spawn(actor.PropsFromProducer(createChildActor))
	default:
		blog.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
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
	system.Root.Send(parentPID, actor.WrapEnvelope(&MessageCreateChild{}))

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
