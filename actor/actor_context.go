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
	ac.receiveTimeoutTimer = time.AfterFunc(d, ac.receiveTimeoutHandler)
}

func (ac *actorContext) receiveTimeoutHandler() {
	ac.CancelReceiveTimeout()
	ac.Send(ac.self, makeMessage[ReceiveTimeout]())
}

func (ac *actorContext) CancelReceiveTimeout() {
	if ac.receiveTimeoutTimer == nil {
		return
	}
	ac.receiveTimeout = 0
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

func (ac *actorContext) Message() *MessageEnvelope {
	return ac.envelope
}

func (ac *actorContext) MessageHeader() ReadonlyMessageHeader {
	if ac.envelope == nil {
		return nil
	}
	return ac.envelope.Header
}

func (ac *actorContext) sendUserMessage(pid *PID, envelope *MessageEnvelope) {
	pid.sendUserMessage(ac.actorSystem, envelope)
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

	pid, err = props.spawn(ac.actorSystem, ac.self.ID+"/"+name, ac)

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
	pid.sendUserMessage(ac.actorSystem, makeMessage[PoisonPill]())
}

//
// Interface: stopperPart
//

func (ac *actorContext) incarnateActor() {
	atomic.StoreInt32(&ac.state, stateAlive)
	ac.actor = ac.props.producer()
}

func (ac *actorContext) InvokeSystemMessage(message *MessageEnvelope) {
	switch message.Message.(type) {
	case *Started:
		ac.InvokeUserMessage(message)
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

func (ac *actorContext) InvokeUserMessage(message *MessageEnvelope) {
	//TODO implement me
	panic("implement me")
}

// I am stopping.
func (ac *actorContext) handleStop() {
	if atomic.LoadInt32(&ac.state) >= stateStopping {
		// already stopping or stopped
		return
	}

	atomic.StoreInt32(&ac.state, stateStopping)

	ac.InvokeUserMessage(makeMessage[Stopping]())
	ac.stopAllChildren()
	ac.tryRestartOrTerminate()
}

// child stopped, check if we can stop or restart (if needed).
func (ac *actorContext) handleTerminated(message *MessageEnvelope) {
	terminated, _ := message.Message.(*Terminated)
	ac.children.Remove(terminated.Who)

	ac.InvokeUserMessage(message)
	ac.tryRestartOrTerminate()
}

func (ac *actorContext) handleRestart() {
	atomic.StoreInt32(&ac.state, stateRestarting)
	ac.InvokeUserMessage(makeMessage[Restarting]())
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
	ac.self.sendSystemMessage(ac.actorSystem, makeMessage[ResumeMailbox]())
	ac.InvokeUserMessage(makeMessage[Started]())

	//if ctx.extras != nil && ctx.extras.stash != nil {
	//	for !ctx.extras.stash.Empty() {
	//		msg, _ := ctx.extras.stash.Pop()
	//		ctx.InvokeUserMessage(msg)
	//	}
	//}
}

func (ac *actorContext) finalizeStop() {
	ac.actorSystem.ProcessRegistry.Remove(ac.self)
	ac.InvokeUserMessage(makeMessage[Stopped]())

	otherStopped := &MessageEnvelope{
		Header:  nil,
		Message: &Terminated{Who: ac.self},
		Sender:  nil,
	}
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
