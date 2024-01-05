package actor

import (
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/colin1989/battery/logger"
)

const (
	stateAlive int32 = iota
	stateRestarting
	stateStopping
	stateStopped
)

type actorContext struct {
	actor               Actor
	actorSystem         *ActorSystem
	extras              *actorContextExtras
	props               *Props
	parent              *PID
	self                *PID
	receiveTimeout      time.Duration
	receiveTimeoutTimer *time.Timer
	envelope            *MessageEnvelope
	state               int32
}

var (
	_ SenderContext   = &actorContext{}
	_ ReceiverContext = &actorContext{}
	_ SpawnerContext  = &actorContext{}
	_ basePart        = &actorContext{}
	_ stopperPart     = &actorContext{}
	_ Invoker         = &actorContext{}
)

func newActorContext(actorSystem *ActorSystem, props *Props, parent *PID) *actorContext {
	ac := &actorContext{
		actorSystem: actorSystem,
		props:       props,
		parent:      parent,
	}
	ac.incarnateActor()
	return ac
}

func (ctx *actorContext) ensureExtras() *actorContextExtras {
	if ctx.extras == nil {
		ctxd := Context(ctx)
		if ctx.props != nil && ctx.props.contextDecoratorChain != nil {
			ctxd = ctx.props.contextDecoratorChain(ctxd)
		}

		ctx.extras = newActorContextExtras(ctxd)
	}

	return ctx.extras
}

//
// Interface: basePart
//

func (ctx *actorContext) Logger() *slog.Logger {
	return ctx.actorSystem.Logger()
}

func (ctx *actorContext) Children() []*PID {
	if ctx.extras == nil {
		return make([]*PID, 0)
	}

	return ctx.extras.Children()
}

func (ctx *actorContext) Respond(response *MessageEnvelope) {
	if ctx.Sender() == nil {
		ctx.actorSystem.DeadLetter.SendUserMessage(nil, response)
		return
	}

	ctx.Send(ctx.Sender(), response)
}

func (ctx *actorContext) Stash() {
	ctx.ensureExtras().stash(ctx.Message())
}

func (ctx *actorContext) Watch(who *PID) {
	who.sendSystemMessage(ctx.actorSystem, &Watch{
		Watcher: ctx.self,
	})
}

func (ctx *actorContext) Unwatch(who *PID) {
	who.sendSystemMessage(ctx.actorSystem, &Unwatch{
		Watcher: ctx.self,
	})
}

func (ctx *actorContext) ReceiveTimeout() time.Duration {
	return ctx.receiveTimeout
}

func (ctx *actorContext) SetReceiveTimeout(d time.Duration) {
	if d <= 0 {
		panic("Duration must be greater than zero")
	}

	if d < time.Millisecond {
		d = 0
	}

	if d == ctx.receiveTimeout {
		return
	}

	ctx.receiveTimeout = d

	ctx.ensureExtras()
	ctx.extras.stopReceiveTimeoutTimer()

	if d > 0 {
		if ctx.extras.receiveTimeoutTimer == nil {
			ctx.extras.initReceiveTimeoutTimer(time.AfterFunc(d, ctx.receiveTimeoutHandler))
		} else {
			ctx.extras.resetReceiveTimeoutTimer(d)
		}
	}
}

func (ctx *actorContext) CancelReceiveTimeout() {
	if ctx.extras == nil || ctx.extras.receiveTimeoutTimer == nil {
		return
	}

	ctx.extras.killReceiveTimeoutTimer()
	ctx.receiveTimeout = 0
}

func (ctx *actorContext) receiveTimeoutHandler() {
	if ctx.extras != nil && ctx.extras.receiveTimeoutTimer != nil {
		ctx.CancelReceiveTimeout()
		//ac.Send(ac.self, receiveTimeoutMessage())
		ctx.self.sendSystemMessage(ctx.actorSystem, receiveTimeoutMessage)
	}
}

//
// Interface: SenderContext
//

func (ctx *actorContext) Parent() *PID {
	return ctx.parent
}

func (ctx *actorContext) Self() *PID {
	return ctx.self
}

func (ctx *actorContext) Actor() Actor {
	return ctx.actor
}

func (ctx *actorContext) ActorSystem() *ActorSystem {
	return ctx.actorSystem
}

func (ctx *actorContext) Sender() *PID {
	if ctx.envelope == nil {
		return nil
	}
	return ctx.envelope.Sender
}

func (ctx *actorContext) Send(pid *PID, envelope *MessageEnvelope) {
	envelope.Sender = ctx.self
	ctx.sendUserMessage(pid, envelope)
}

func (ctx *actorContext) Request(pid *PID, message interface{}) (*MessageEnvelope, error) {
	// TODO: timeout 应该作为配置
	timeout := time.Second * 5
	future := NewFuture(ctx.actorSystem, timeout)
	envelope := &MessageEnvelope{
		Header:  nil,
		Message: message,
		Sender:  future.pid,
	}
	ctx.sendUserMessage(pid, envelope)
	return future.Result()
}

func (ctx *actorContext) Envelope() *MessageEnvelope {
	return ctx.envelope
}

func (ctx *actorContext) MessageHeader() ReadonlyMessageHeader {
	if ctx.envelope == nil {
		return nil
	}
	return ctx.envelope.Header
}

func (ctx *actorContext) sendUserMessage(pid *PID, envelope *MessageEnvelope) {
	if ctx.props.senderMiddlewareChain != nil {
		ctx.props.senderMiddlewareChain(ctx, pid, envelope)
	} else {
		pid.sendUserMessage(ctx.actorSystem, envelope)
	}
}

func (ctx *actorContext) Message() interface{} {
	return UnwrapEnvelopeMessage(ctx.envelope)
}

//
// Interface: ReceiverContext
//

func (ctx *actorContext) Receive(envelope *MessageEnvelope) {
	ctx.envelope = envelope
	ctx.defaultReceive()
	ctx.envelope = nil
}

func (ctx *actorContext) defaultReceive() {
	switch ctx.envelope.Message.(type) {
	case *PoisonPill:
		ctx.Stop(ctx.self)
	default:
		ctx.actor.Receive(ctx)
	}
}

//
// Interface: SpawnerContext
//

func (ctx *actorContext) Spawn(props *Props) *PID {
	pid, err := ctx.SpawnNamed(props, ctx.actorSystem.ProcessRegistry.NextId())
	if err != nil {
		panic(err)
	}

	return pid
}

func (ctx *actorContext) SpawnPrefix(props *Props, prefix string) *PID {
	pid, err := ctx.SpawnNamed(props, prefix+ctx.actorSystem.ProcessRegistry.NextId())
	if err != nil {
		panic(err)
	}

	return pid
}

func (ctx *actorContext) SpawnNamed(props *Props, name string) (*PID, error) {
	var pid *PID
	var err error

	if ctx.props.spawnMiddlewareChain != nil {
		pid, err = ctx.props.spawnMiddlewareChain(ctx.actorSystem, ctx.self.ID+"/"+name, props, ctx)
	} else {
		pid, err = props.spawn(ctx.actorSystem, ctx.self.ID+"/"+name, ctx)
	}

	if err != nil {
		return pid, err
	}

	ctx.ensureExtras().addChild(pid)

	return pid, err
}

//
// Interface: stopperPart
//

func (ctx *actorContext) Stop(pid *PID) {
	pid.ref(ctx.actorSystem).Stop(pid)
}

// StopFuture will stop actor immediately regardless of existing user messages in mailbox, and return its future.
func (ctx *actorContext) StopFuture(pid *PID) *Future {
	future := NewFuture(ctx.actorSystem, 10*time.Second)

	pid.sendSystemMessage(ctx.actorSystem, &Watch{Watcher: future.pid})
	ctx.Stop(pid)

	return future
}

func (ctx *actorContext) Poison(pid *PID) {
	pid.sendUserMessage(ctx.actorSystem, PoisonPillMessage())
}

// PoisonFuture will tell actor to stop after processing current user messages in mailbox, and return its future.
func (ctx *actorContext) PoisonFuture(pid *PID) *Future {
	future := NewFuture(ctx.actorSystem, 10*time.Second)

	pid.sendSystemMessage(ctx.actorSystem, &Watch{Watcher: future.pid})
	ctx.Poison(pid)

	return future
}

//
// Interface: stopperPart
//

func (ctx *actorContext) incarnateActor() {
	atomic.StoreInt32(&ctx.state, stateAlive)
	ctx.actor = ctx.props.producer(ctx.actorSystem)
}

func (ctx *actorContext) EscalateFailure(reason interface{}, message interface{}) {
	ctx.self.sendSystemMessage(ctx.actorSystem, suspendMailboxMessage)

	failure := &Failure{
		Reason: reason,
		Who:    ctx.self,
		//RestartStats: ctx.ensureExtras().restartStats(),
		Message: message,
	}

	if ctx.parent == nil {
		ctx.handleRootFailure(failure)
	} else {
		ctx.parent.sendSystemMessage(ctx.actorSystem, failure)
	}
}

func (ctx *actorContext) InvokeSystemMessage(message SystemMessage) {
	switch msg := message.(type) {
	case *Started:
		ctx.InvokeUserMessage(startedMessageEnvelope())
	case *Watch:
		ctx.handleWatch(msg)
	case *Unwatch:
		ctx.handleUnwatch(msg)
	case *Stop:
		ctx.handleStop()
	case *Terminated:
		ctx.handleTerminated(msg)
	case *Restart:
		ctx.handleRestart()
	default:
		logger.Warn("unknown system message", slog.Any("message", message))
	}
}

func (ctx *actorContext) handleRootFailure(failure *Failure) {
	//defaultSupervisionStrategy.HandleFailure(ctx.actorSystem, ctx, failure.Who, failure.RestartStats, failure.Reason, failure.Envelope)
	logger.Warn("handleRootFailure", slog.Any("failure", failure))
}

func (ctx *actorContext) InvokeUserMessage(envelope *MessageEnvelope) {
	if atomic.LoadInt32(&ctx.state) == stateStopped {
		return
	}

	_, msg, _ := UnwrapEnvelope(envelope)

	influenceTimeout := true
	if ctx.receiveTimeout > 0 {
		_, influenceTimeout = msg.(NotInfluenceReceiveTimeout)
		influenceTimeout = !influenceTimeout

		if influenceTimeout {
			ctx.extras.stopReceiveTimeoutTimer()
		}
	}

	ctx.processMessage(envelope)

	if ctx.receiveTimeout > 0 && influenceTimeout {
		ctx.extras.resetReceiveTimeoutTimer(ctx.receiveTimeout)
	}
}

func (ctx *actorContext) processMessage(envelope *MessageEnvelope) {
	if ctx.props.receiverMiddlewareChain != nil {
		ctx.props.receiverMiddlewareChain(ctx, envelope)

		return
	}

	ctx.envelope = envelope
	ctx.defaultReceive()
	ctx.envelope = nil
}

// I am stopping.
func (ctx *actorContext) handleStop() {
	if atomic.LoadInt32(&ctx.state) >= stateStopping {
		// already stopping or stopped
		return
	}

	logger.Warn("actor handleStop", slog.String("pid", ctx.self.String()))
	atomic.StoreInt32(&ctx.state, stateStopping)

	ctx.InvokeUserMessage(stoppingMessage())
	ctx.stopAllChildren()
	ctx.tryRestartOrTerminate()
}

func (ctx *actorContext) handleWatch(msg *Watch) {
	if atomic.LoadInt32(&ctx.state) >= stateStopping {
		msg.Watcher.sendSystemMessage(ctx.actorSystem, &Terminated{
			Who: ctx.self,
		})
	} else {
		ctx.ensureExtras().watch(msg.Watcher)
	}
}

func (ctx *actorContext) handleUnwatch(msg *Unwatch) {
	if ctx.extras == nil {
		return
	}

	ctx.extras.unwatch(msg.Watcher)
}

// child stopped, check if we can stop or restart (if needed).
func (ctx *actorContext) handleTerminated(terminated *Terminated) {
	if ctx.extras != nil {
		ctx.extras.removeChild(terminated.Who)
	}
	ctx.InvokeUserMessage(WrapEnvelop(terminated))
	ctx.tryRestartOrTerminate()
}

func (ctx *actorContext) handleRestart() {
	atomic.StoreInt32(&ctx.state, stateRestarting)
	ctx.InvokeUserMessage(restartingMessage())
	ctx.stopAllChildren()
	ctx.tryRestartOrTerminate()
}

func (ctx *actorContext) stopAllChildren() {
	if ctx.extras == nil {
		return
	}

	pids := ctx.extras.Children()
	for i := len(pids) - 1; i >= 0; i-- {
		pids[i].sendSystemMessage(ctx.actorSystem, stopMessage)
	}
}

func (ctx *actorContext) tryRestartOrTerminate() {
	if ctx.extras != nil && !ctx.extras.children.Empty() {
		return
	}

	switch atomic.LoadInt32(&ctx.state) {
	case stateRestarting:
		ctx.CancelReceiveTimeout()
		ctx.restart()
	case stateStopping:
		ctx.CancelReceiveTimeout()
		ctx.finalizeStop()
	}
}

func (ctx *actorContext) restart() {
	ctx.incarnateActor()
	ctx.self.sendSystemMessage(ctx.actorSystem, resumeMailboxMessage)
	ctx.InvokeUserMessage(startedMessageEnvelope())

	for {
		msg, ok := ctx.extras.popStash()
		if !ok {
			break
		}
		ctx.InvokeUserMessage(WrapEnvelop(msg))
	}
}

func (ctx *actorContext) finalizeStop() {
	ctx.actorSystem.ProcessRegistry.Remove(ctx.self)
	ctx.InvokeUserMessage(stoppedMessage())

	otherStopped := &Terminated{Who: ctx.self}
	// Notify watchers
	if ctx.extras != nil {
		ctx.extras.watchers.ForEach(func(i int, pid *PID) {
			pid.sendSystemMessage(ctx.actorSystem, otherStopped)
		})
	}
	// Notify parent
	if ctx.parent != nil {
		ctx.parent.sendSystemMessage(ctx.actorSystem, otherStopped)
	}

	atomic.StoreInt32(&ctx.state, stateStopped)
}
