package actor

import (
	"log/slog"
	"time"
)

type RootContext struct {
	actorSystem      *ActorSystem
	senderMiddleware SenderFunc
	spawnMiddleware  SpawnFunc
	headers          messageHeader
}

var (
	_ SenderContext  = &RootContext{}
	_ SpawnerContext = &RootContext{}
	_ stopperPart    = &RootContext{}
)

func NewRootContext(actorSystem *ActorSystem, header map[string]string, middleware ...SenderMiddleware) *RootContext {
	if header == nil {
		header = make(map[string]string)
	}

	rc := &RootContext{
		actorSystem: actorSystem,
		senderMiddleware: makeSenderMiddlewareChain(middleware, func(_ SenderContext, target *PID, envelope *MessageEnvelope) {
			target.sendUserMessage(actorSystem, envelope)
		}),
		headers: header,
	}

	return rc
}

func (rc *RootContext) ActorSystem() *ActorSystem {
	return rc.actorSystem
}

func (rc *RootContext) Logger() *slog.Logger {
	return rc.actorSystem.Logger()
}

func (rc *RootContext) WithHeaders(headers map[string]string) *RootContext {
	rc.headers = headers

	return rc
}

func (rc *RootContext) WithSenderMiddleware(middleware ...SenderMiddleware) *RootContext {
	rc.senderMiddleware = makeSenderMiddlewareChain(middleware, func(_ SenderContext, target *PID, envelope *MessageEnvelope) {
		target.sendUserMessage(rc.actorSystem, envelope)
	})

	return rc
}

func (rc *RootContext) WithSpawnMiddleware(middleware ...SpawnMiddleware) *RootContext {
	rc.spawnMiddleware = makeSpawnMiddlewareChain(middleware, func(actorSystem *ActorSystem, id string, props *Props, parentContext SpawnerContext) (pid *PID, e error) {
		return props.spawn(actorSystem, id, rc)
	})

	return rc
}

//
// Interface: info
//

func (rc *RootContext) Parent() *PID {
	return nil
}

func (rc *RootContext) Self() *PID {
	return nil
}

func (rc *RootContext) Sender() *PID {
	return nil
}

func (rc *RootContext) Actor() Actor {
	return nil
}

//
// Interface: sender
//

func (rc *RootContext) Send(pid *PID, envelope *MessageEnvelope) {
	if rc.senderMiddleware != nil {
		// Request based middleware
		rc.senderMiddleware(rc, pid, envelope)
	} else {
		// tell based middleware
		pid.sendUserMessage(rc.actorSystem, envelope)
	}
}

// Request
//
//	@Description:
//	@receiver rc
//	@param pid
//	@param message message's type cannot be MessageEnvelope
//	@return *MessageEnvelope
//	@return error
func (rc *RootContext) Request(pid *PID, envelope *MessageEnvelope) (*MessageEnvelope, error) {
	// TODO: timeout 应该作为配置
	timeout := time.Second * 5
	future := NewFuture(rc.actorSystem, timeout)
	envelope.Sender = future.pid

	pid.sendUserMessage(rc.actorSystem, envelope)
	return future.Result()
}

//
// Interface: message
//

func (rc *RootContext) Envelope() *MessageEnvelope {
	return nil
}

func (rc *RootContext) MessageHeader() ReadonlyMessageHeader {
	return rc.headers
}

func (rc *RootContext) Stop(pid *PID) {
	pid.ref(rc.actorSystem).Stop(pid)
}

// StopFuture will stop actor immediately regardless of existing user messages in mailbox, and return its future.
func (rc *RootContext) StopFuture(pid *PID) *Future {
	future := NewFuture(rc.actorSystem, 10*time.Second)

	pid.sendSystemMessage(rc.actorSystem, &Watch{Watcher: future.pid})
	rc.Stop(pid)

	return future
}

func (rc *RootContext) Poison(pid *PID) {
	pid.sendUserMessage(rc.actorSystem, PoisonPillMessage())
}

// PoisonFuture will tell actor to stop after processing current user messages in mailbox, and return its future.
func (rc *RootContext) PoisonFuture(pid *PID) *Future {
	future := NewFuture(rc.actorSystem, 10*time.Second)

	pid.sendSystemMessage(rc.actorSystem, &Watch{Watcher: future.pid})
	rc.Poison(pid)

	return future
}

//
// Interface: Spawn
//

func (rc *RootContext) Spawn(props *Props) *PID {
	pid, err := rc.SpawnNamed(props, rc.actorSystem.ProcessRegistry.NextId())
	if err != nil {
		panic(err)
	}

	return pid
}

func (rc *RootContext) SpawnPrefix(props *Props, prefix string) *PID {
	pid, err := rc.SpawnNamed(props, prefix+rc.actorSystem.ProcessRegistry.NextId())
	if err != nil {
		panic(err)
	}

	return pid
}

func (rc *RootContext) SpawnNamed(props *Props, name string) (*PID, error) {
	if rc.spawnMiddleware != nil {
		return rc.spawnMiddleware(rc.actorSystem, name, props, rc)
	}

	return props.spawn(rc.actorSystem, name, rc)
}
