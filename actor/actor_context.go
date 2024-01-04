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

func (ac *actorContext) ensureExtras() *actorContextExtras {
	if ac.extras == nil {
		ctxd := Context(ac)
		if ac.props != nil && ac.props.contextDecoratorChain != nil {
			ctxd = ac.props.contextDecoratorChain(ctxd)
		}

		ac.extras = newActorContextExtras(ctxd)
	}

	return ac.extras
}

//
// Interface: basePart
//

func (ac *actorContext) Children() []*PID {
	if ac.extras == nil {
		return make([]*PID, 0)
	}

	return ac.extras.Children()
}

func (ac *actorContext) Respond(response *MessageEnvelope) {
	if ac.Sender() == nil {
		ac.actorSystem.DeadLetter.SendUserMessage(nil, response)
		return
	}

	ac.Send(ac.Sender(), response)
}

func (ac *actorContext) Stash() {
	ac.ensureExtras().stash(ac.Message())
}

func (ac *actorContext) Watch(who *PID) {
	who.sendSystemMessage(ac.actorSystem, &Watch{
		Watcher: ac.self,
	})
}

func (ac *actorContext) Unwatch(who *PID) {
	who.sendSystemMessage(ac.actorSystem, &Unwatch{
		Watcher: ac.self,
	})
}

func (ac *actorContext) ReceiveTimeout() time.Duration {
	return ac.receiveTimeout
}

func (ac *actorContext) SetReceiveTimeout(d time.Duration) {
	if d <= 0 {
		panic("Duration must be greater than zero")
	}

	if d < time.Millisecond {
		d = 0
	}

	if d == ac.receiveTimeout {
		return
	}

	ac.receiveTimeout = d

	ac.ensureExtras()
	ac.extras.stopReceiveTimeoutTimer()

	if d > 0 {
		if ac.extras.receiveTimeoutTimer == nil {
			ac.extras.initReceiveTimeoutTimer(time.AfterFunc(d, ac.receiveTimeoutHandler))
		} else {
			ac.extras.resetReceiveTimeoutTimer(d)
		}
	}
}

func (ac *actorContext) CancelReceiveTimeout() {
	if ac.extras == nil || ac.extras.receiveTimeoutTimer == nil {
		return
	}

	ac.extras.killReceiveTimeoutTimer()
	ac.receiveTimeout = 0
}

func (ac *actorContext) receiveTimeoutHandler() {
	if ac.extras != nil && ac.extras.receiveTimeoutTimer != nil {
		ac.CancelReceiveTimeout()
		//ac.Send(ac.self, receiveTimeoutMessage())
		ac.self.sendSystemMessage(ac.actorSystem, receiveTimeoutMessage)
	}
}

//
// Interface: SenderContext
//

func (ac *actorContext) Parent() *PID {
	return ac.parent
}

func (ac *actorContext) Self() *PID {
	return ac.self
}

func (ac *actorContext) Actor() Actor {
	return ac.actor
}

func (ac *actorContext) ActorSystem() *ActorSystem {
	return ac.actorSystem
}

func (ac *actorContext) Sender() *PID {
	if ac.envelope == nil {
		return nil
	}
	return ac.envelope.Sender
}

func (ac *actorContext) Send(pid *PID, envelope *MessageEnvelope) {
	envelope.Sender = ac.self
	ac.sendUserMessage(pid, envelope)
}

func (ac *actorContext) Request(pid *PID, message interface{}) (*MessageEnvelope, error) {
	// TODO: timeout 应该作为配置
	timeout := time.Second * 5
	future := NewFuture(ac.actorSystem, timeout)
	envelope := &MessageEnvelope{
		Header:  nil,
		Message: message,
		Sender:  future.pid,
	}
	ac.sendUserMessage(pid, envelope)
	return future.Result()
}

func (ac *actorContext) Envelope() *MessageEnvelope {
	return ac.envelope
}

func (ac *actorContext) MessageHeader() ReadonlyMessageHeader {
	if ac.envelope == nil {
		return nil
	}
	return ac.envelope.Header
}

func (ac *actorContext) sendUserMessage(pid *PID, envelope *MessageEnvelope) {
	if ac.props.senderMiddlewareChain != nil {
		ac.props.senderMiddlewareChain(ac, pid, envelope)
	} else {
		pid.sendUserMessage(ac.actorSystem, envelope)
	}
}

func (ac *actorContext) Message() interface{} {
	return UnwrapEnvelopeMessage(ac.envelope)
}

//
// Interface: ReceiverContext
//

func (ac *actorContext) Receive(envelope *MessageEnvelope) {
	ac.envelope = envelope
	ac.defaultReceive()
	ac.envelope = nil
}

func (ac *actorContext) defaultReceive() {
	switch ac.envelope.Message.(type) {
	case *PoisonPill:
		ac.Stop(ac.self)
	default:
		ac.actor.Receive(ac)
	}
}

//
// Interface: SpawnerContext
//

func (ac *actorContext) Spawn(props *Props) *PID {
	pid, err := ac.SpawnNamed(props, ac.actorSystem.ProcessRegistry.NextId())
	if err != nil {
		panic(err)
	}

	return pid
}

func (ac *actorContext) SpawnPrefix(props *Props, prefix string) *PID {
	pid, err := ac.SpawnNamed(props, prefix+ac.actorSystem.ProcessRegistry.NextId())
	if err != nil {
		panic(err)
	}

	return pid
}

func (ac *actorContext) SpawnNamed(props *Props, name string) (*PID, error) {
	var pid *PID
	var err error

	if ac.props.spawnMiddlewareChain != nil {
		pid, err = ac.props.spawnMiddlewareChain(ac.actorSystem, ac.self.ID+"/"+name, props, ac)
	} else {
		pid, err = props.spawn(ac.actorSystem, ac.self.ID+"/"+name, ac)
	}

	if err != nil {
		return pid, err
	}

	ac.ensureExtras().addChild(pid)

	return pid, err
}

//
// Interface: stopperPart
//

func (ac *actorContext) Stop(pid *PID) {
	pid.ref(ac.actorSystem).Stop(pid)
}

func (ac *actorContext) Poison(pid *PID) {
	pid.sendUserMessage(ac.actorSystem, poisonPillMessage())
}

//
// Interface: stopperPart
//

func (ac *actorContext) incarnateActor() {
	atomic.StoreInt32(&ac.state, stateAlive)
	ac.actor = ac.props.producer()
}

func (ac *actorContext) EscalateFailure(reason interface{}, message interface{}) {
	ac.self.sendSystemMessage(ac.actorSystem, suspendMailboxMessage)

	failure := &Failure{
		Reason: reason,
		Who:    ac.self,
		//RestartStats: ctx.ensureExtras().restartStats(),
		Message: message,
	}

	if ac.parent == nil {
		ac.handleRootFailure(failure)
	} else {
		ac.parent.sendSystemMessage(ac.actorSystem, failure)
	}
}

func (ac *actorContext) InvokeSystemMessage(message SystemMessage) {
	switch msg := message.(type) {
	case *Started:
		ac.InvokeUserMessage(startedMessageEnvelope())
	case *Watch:
		ac.handleWatch(msg)
	case *Unwatch:
		ac.handleUnwatch(msg)
	case *Stop:
		ac.handleStop()
	case *Terminated:
		ac.handleTerminated(msg)
	case *Restart:
		ac.handleRestart()
	default:
		logger.Warn("unknown system message", slog.Any("message", message))
	}
}

func (ac *actorContext) handleRootFailure(failure *Failure) {
	//defaultSupervisionStrategy.HandleFailure(ctx.actorSystem, ctx, failure.Who, failure.RestartStats, failure.Reason, failure.Envelope)
	logger.Warn("handleRootFailure", slog.Any("failure", failure))
}

func (ac *actorContext) InvokeUserMessage(envelope *MessageEnvelope) {
	if atomic.LoadInt32(&ac.state) == stateStopped {
		return
	}

	_, msg, _ := UnwrapEnvelope(envelope)

	influenceTimeout := true
	if ac.receiveTimeout > 0 {
		_, influenceTimeout = msg.(NotInfluenceReceiveTimeout)
		influenceTimeout = !influenceTimeout

		if influenceTimeout {
			ac.extras.stopReceiveTimeoutTimer()
		}
	}

	ac.processMessage(envelope)

	if ac.receiveTimeout > 0 && influenceTimeout {
		ac.extras.resetReceiveTimeoutTimer(ac.receiveTimeout)
	}
}

func (ac *actorContext) processMessage(envelope *MessageEnvelope) {
	if ac.props.receiverMiddlewareChain != nil {
		ac.props.receiverMiddlewareChain(ac, envelope)

		return
	}

	ac.envelope = envelope
	ac.defaultReceive()
	ac.envelope = nil
}

// I am stopping.
func (ac *actorContext) handleStop() {
	if atomic.LoadInt32(&ac.state) >= stateStopping {
		// already stopping or stopped
		return
	}

	logger.Warn("actor handleStop", slog.String("pid", ac.self.String()))
	atomic.StoreInt32(&ac.state, stateStopping)

	ac.InvokeUserMessage(stoppingMessage())
	ac.stopAllChildren()
	ac.tryRestartOrTerminate()
}

func (ac *actorContext) handleWatch(msg *Watch) {
	if atomic.LoadInt32(&ac.state) >= stateStopping {
		msg.Watcher.sendSystemMessage(ac.actorSystem, &Terminated{
			Who: ac.self,
		})
	} else {
		ac.ensureExtras().watch(msg.Watcher)
	}
}

func (ac *actorContext) handleUnwatch(msg *Unwatch) {
	if ac.extras == nil {
		return
	}

	ac.extras.unwatch(msg.Watcher)
}

// child stopped, check if we can stop or restart (if needed).
func (ac *actorContext) handleTerminated(terminated *Terminated) {
	if ac.extras != nil {
		ac.extras.removeChild(terminated.Who)
	}
	ac.InvokeUserMessage(WrapEnvelop(terminated))
	ac.tryRestartOrTerminate()
}

func (ac *actorContext) handleRestart() {
	atomic.StoreInt32(&ac.state, stateRestarting)
	ac.InvokeUserMessage(restartingMessage())
	ac.stopAllChildren()
	ac.tryRestartOrTerminate()
}

func (ac *actorContext) stopAllChildren() {
	if ac.extras == nil {
		return
	}

	pids := ac.extras.Children()
	for i := len(pids) - 1; i >= 0; i-- {
		pids[i].sendSystemMessage(ac.actorSystem, stopMessage)
	}
}

func (ac *actorContext) tryRestartOrTerminate() {
	if ac.extras != nil && !ac.extras.children.Empty() {
		return
	}

	switch atomic.LoadInt32(&ac.state) {
	case stateRestarting:
		ac.CancelReceiveTimeout()
		ac.restart()
	case stateStopping:
		ac.CancelReceiveTimeout()
		ac.finalizeStop()
	}
}

func (ac *actorContext) restart() {
	ac.incarnateActor()
	ac.self.sendSystemMessage(ac.actorSystem, resumeMailboxMessage)
	ac.InvokeUserMessage(startedMessageEnvelope())

	for {
		msg, ok := ac.extras.popStash()
		if !ok {
			break
		}
		ac.InvokeUserMessage(WrapEnvelop(msg))
	}
}

func (ac *actorContext) finalizeStop() {
	ac.actorSystem.ProcessRegistry.Remove(ac.self)
	ac.InvokeUserMessage(stoppedMessage())

	otherStopped := &Terminated{Who: ac.self}
	// Notify watchers
	if ac.extras != nil {
		ac.extras.watchers.ForEach(func(i int, pid *PID) {
			pid.sendSystemMessage(ac.actorSystem, otherStopped)
		})
	}
	// Notify parent
	if ac.parent != nil {
		ac.parent.sendSystemMessage(ac.actorSystem, otherStopped)
	}

	atomic.StoreInt32(&ac.state, stateStopped)
}
