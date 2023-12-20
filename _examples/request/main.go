package main

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/logger"
	"log/slog"
	"reflect"
	"sync"
)

// addition, subtraction, multiplication and division
type (
	mathActor struct{}
	addActor  struct{}
	subActor  struct{}
	mulActor  struct{}
	divActor  struct{}
	addition  struct {
		A, B float64
	}
	subtraction struct {
		A, B float64
	}
	multiplication struct {
		A, B float64
	}
	division struct {
		A, B float64
	}
	out struct {
		Value float64
	}
)

func (a *addActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	var result out
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		return
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		return
	case *addition:
		result.Value = msg.A + msg.B
	default:
		logger.Warn("addActor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
		return
	}
	ctx.Respond(actor.WrapEnvelop(result))
}

func (s *subActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	var result out
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		return
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		return
	case *subtraction:
		result.Value = msg.A - msg.B
	default:
		logger.Warn("subActor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
		return
	}
	ctx.Respond(actor.WrapEnvelop(result))
}

func (m *mulActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	var result out
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		return
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		return
	case *multiplication:
		result.Value = msg.A * msg.B
	default:
		logger.Warn("mulActor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
		return
	}
	ctx.Respond(actor.WrapEnvelop(result))
}

func (d *divActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	var result out
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		return
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		return
	case *division:
		result.Value = msg.A / msg.B
	default:
		logger.Warn("divActor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
		return
	}
	ctx.Respond(actor.WrapEnvelop(result))
}

func (m *mathActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	var result interface{}
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		return
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		return
	case *addition:
		if addPID == nil {
			addPID = ctx.Spawn(actor.PropsFromProducer(func() actor.Actor {
				return &addActor{}
			}))
		}
		request, err := ctx.Request(addPID, msg)
		if err != nil {
			logger.Error("Request addition", logger.ErrAttr(err))
			return
		}
		result = request.Message
	case *subtraction:
		if subPID == nil {
			subPID = ctx.Spawn(actor.PropsFromProducer(func() actor.Actor {
				return &subActor{}
			}))
		}
		request, err := ctx.Request(subPID, msg)
		if err != nil {
			logger.Error("Request subtraction", logger.ErrAttr(err))
			return
		}
		result = request.Message
	case *multiplication:
		if mulPID == nil {
			mulPID = ctx.Spawn(actor.PropsFromProducer(func() actor.Actor {
				return &mulActor{}
			}))
		}
		request, err := ctx.Request(mulPID, msg)
		if err != nil {
			logger.Error("Request multiplication", logger.ErrAttr(err))
			return
		}
		result = request.Message
	case *division:
		if divPID == nil {
			divPID = ctx.Spawn(actor.PropsFromProducer(func() actor.Actor {
				return &divActor{}
			}))
		}
		request, err := ctx.Request(divPID, msg)
		if err != nil {
			logger.Error("Request division", logger.ErrAttr(err))
			return
		}
		result = request.Message
	default:
		logger.Warn("mathActor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
		return
	}
	ctx.Respond(actor.WrapEnvelop(result))
}

func requestAddition(m *actor.PID, a, b float64) {
	defer func() {
		wg.Done()
	}()
	add := &addition{
		A: a,
		B: b,
	}
	result, err := system.Root.Request(m, add)
	if err != nil {
		logger.Error("Request addition", logger.ErrAttr(err))
		return
	}
	if result.Message.(out).Value != a+b {
		panic("requestAddition does not equal")
	}
	logger.Info(fmt.Sprintf(" Request addition[%v]=[%v] \n", add, result))
}

func requestSubtraction(m *actor.PID, a, b float64) {
	defer func() {
		wg.Done()
	}()
	sub := &subtraction{
		A: a,
		B: b,
	}
	result, err := system.Root.Request(m, sub)
	if err != nil {
		logger.Error("Request subtraction", logger.ErrAttr(err))
		return
	}
	if result.Message.(out).Value != a-b {
		panic("requestSubtraction does not equal")
	}
	logger.Info(fmt.Sprintf(" Request subtraction[%v]=[%v] \n", sub, result))
}

func requestMultiplication(m *actor.PID, a, b float64) {
	defer func() {
		wg.Done()
	}()
	mul := &multiplication{
		A: a,
		B: b,
	}
	result, err := system.Root.Request(m, mul)
	if err != nil {
		logger.Error("Request multiplication", logger.ErrAttr(err))
		return
	}
	if result.Message.(out).Value != a*b {
		panic("requestMultiplication does not equal")
	}
	logger.Info(fmt.Sprintf(" Request multiplication[%v]=[%v] \n", mul, result))
}

func requestDivision(m *actor.PID, a, b float64) {
	defer func() {
		wg.Done()
	}()
	div := &division{
		A: a,
		B: b,
	}
	result, err := system.Root.Request(m, div)
	if err != nil {
		logger.Error("Request division", logger.ErrAttr(err))
		return
	}
	if result.Message.(out).Value != a/b {
		panic("requestDivision does not equal")
	}
	logger.Info(fmt.Sprintf(" Request division[%v]=[%v] \n", div, result))
}

var (
	wg     sync.WaitGroup
	system *actor.ActorSystem
	addPID *actor.PID
	subPID *actor.PID
	mulPID *actor.PID
	divPID *actor.PID
)

func main() {
	system = actor.NewActorSystem()
	props := actor.PropsFromProducer(func() actor.Actor {
		return &mathActor{}
	})
	m := system.Root.Spawn(props)
	count := 1000
	wg.Add(count)
	for i := 0; i < count; i++ {
		switch i % 4 {
		case 0:
			go requestAddition(m, float64(i), float64(i-1))
		case 1:
			go requestSubtraction(m, float64(i), float64(i-1))
		case 2:
			go requestMultiplication(m, float64(i), float64(i/2))
		case 3:
			go requestDivision(m, float64(i), float64(i/2))
		}
	}

	wg.Wait()

	system.Root.Poison(m)
	system.Shutdown()
}
