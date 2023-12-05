package actor

import (
	"fmt"
	"sync/atomic"
	"time"
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
	props               *Props
	parent              *PID
	self                *PID
	receiveTimeout      time.Duration
	receiveTimeoutTimer *time.Timer
	envelope            *MessageEnvelope
	state               int32
	children            PIDSet
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

//
// Interface: basePart
//

func (ac *actorContext) Children() []*PID {
	return ac.children.Values()
}

func (ac *actorContext) Respond(response *MessageEnvelope) {
	if ac.Sender() == nil {
		ac.actorSystem.DeadLetter.SendUserMessage(nil, response)
		return
	}

	ac.Send(ac.Sender(), response)
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

	ac.stopReceiveTimeoutTimer()

	ac.receiveTimeout = d
	if ac.receiveTimeoutTimer == nil {
		ac.initReceiveTimeoutTimer(time.AfterFunc(d, ac.receiveTimeoutHandler))
	} else {
		ac.resetReceiveTimeoutTimer(d)
	}
}

func (ac *actorContext) CancelReceiveTimeout() {
	if ac.receiveTimeoutTimer == nil {
		return
	}

	ac.killReceiveTimeoutTimer()
	ac.receiveTimeout = 0
}

func (ac *actorContext) initReceiveTimeoutTimer(timer *time.Timer) {
	ac.receiveTimeoutTimer = timer
}

func (ac *actorContext) resetReceiveTimeoutTimer(time time.Duration) {
	if ac.receiveTimeoutTimer == nil {
		return
	}

	ac.receiveTimeoutTimer.Reset(time)
}

func (ac *actorContext) receiveTimeoutHandler() {
	ac.CancelReceiveTimeout()
	ac.self.sendSystemMessage(ac.actorSystem, receiveTimeoutMessage)
}

func (ac *actorContext) stopReceiveTimeoutTimer() {
	if ac.receiveTimeoutTimer == nil {
		return
	}

	ac.receiveTimeoutTimer.Stop()
}

func (ac *actorContext) killReceiveTimeoutTimer() {
	if ac.receiveTimeoutTimer == nil {
		return
	}

	ac.receiveTimeoutTimer.Stop()
	ac.receiveTimeoutTimer = nil
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

	ac.children.Add(pid)

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
	switch message.(type) {
	case *Started:
		ac.InvokeUserMessage(startedMessageEnvelope())
	case *Stop:
		ac.handleStop()
	case *Terminated:
		ac.handleTerminated(message)
	case *Restart:
		ac.handleRestart()
	default:
		fmt.Printf("unknown system message %v", message)
	}
}

func (ac *actorContext) handleRootFailure(failure *Failure) {
	//defaultSupervisionStrategy.HandleFailure(ctx.actorSystem, ctx, failure.Who, failure.RestartStats, failure.Reason, failure.Envelope)
	fmt.Printf("handleRootFailure ： %+v", failure)
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
			ac.stopReceiveTimeoutTimer()
		}
	}

	ac.processMessage(envelope)

	if ac.receiveTimeout > 0 && influenceTimeout {
		ac.resetReceiveTimeoutTimer(ac.receiveTimeout)
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

	fmt.Printf("actor[%v] stopping \n", ac.self)
	atomic.StoreInt32(&ac.state, stateStopping)

	ac.InvokeUserMessage(stoppingMessage())
	ac.stopAllChildren()
	ac.tryRestartOrTerminate()
}

// child stopped, check if we can stop or restart (if needed).
func (ac *actorContext) handleTerminated(message SystemMessage) {
	terminated, _ := message.(*Terminated)
	ac.children.Remove(terminated.Who)

	//ac.InvokeUserMessage(message)
	ac.tryRestartOrTerminate()
}

func (ac *actorContext) handleRestart() {
	atomic.StoreInt32(&ac.state, stateRestarting)
	ac.InvokeUserMessage(restartingMessage())
	ac.stopAllChildren()
	ac.tryRestartOrTerminate()
}

func (ac *actorContext) stopAllChildren() {

	ac.children.ForEach(func(_ int, pid *PID) {
		ac.Stop(pid)
	})
}

func (ac *actorContext) tryRestartOrTerminate() {
	if !ac.children.Empty() {
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

	//if ctx.extras != nil && ctx.extras.stash != nil {
	//	for !ctx.extras.stash.Empty() {
	//		msg, _ := ctx.extras.stash.Pop()
	//		ctx.InvokeUserMessage(msg)
	//	}
	//}
}

func (ac *actorContext) finalizeStop() {
	ac.actorSystem.ProcessRegistry.Remove(ac.self)
	ac.InvokeUserMessage(stoppedMessage())

	otherStopped := &Terminated{Who: ac.self}
	//// Notify watchers
	//if ctx.extras != nil {
	//	ctx.extras.watchers.ForEach(func(i int, pid *PID) {
	//		pid.sendSystemMessage(ctx.actorSystem, otherStopped)
	//	})
	//}
	// Notify parent
	if ac.parent != nil {
		ac.parent.sendSystemMessage(ac.actorSystem, otherStopped)
	}

	atomic.StoreInt32(&ac.state, stateStopped)
}
