package router

import (
	"sync"
	"sync/atomic"

	"github.com/colin1989/battery/actor"
)

// routerProcess serves as a proxy to the router implementation and forwards messages directly to the routee. This
// optimization avoids serializing router messages through an actor
type routerProcess struct {
	parent      *actor.PID
	router      *actor.PID
	state       State
	mu          sync.Mutex
	watchers    actor.PIDSet
	stopping    int32
	actorSystem *actor.ActorSystem
}

var _ actor.Process = &routerProcess{}

func (ref *routerProcess) SendUserMessage(pid *actor.PID, envelope *actor.MessageEnvelope) {
	msg := envelope.Message

	// Add support for PoisonPill. Originally only Stop is supported.
	if _, ok := msg.(*actor.PoisonPill); ok {
		ref.Poison(pid)
		return
	}
	if _, ok := msg.(ManagementMessage); !ok {
		ref.state.RouteMessage(envelope)
	} else {
		r, _ := ref.actorSystem.ProcessRegistry.Get(ref.router)
		r.SendUserMessage(pid, envelope)
	}
}

func (ref *routerProcess) SendSystemMessage(pid *actor.PID, message actor.SystemMessage) {
	switch msg := message.(type) {
	case *actor.Watch:
		if atomic.LoadInt32(&ref.stopping) == 1 {
			if r, ok := ref.actorSystem.ProcessRegistry.Get(msg.Watcher); ok {
				r.SendSystemMessage(msg.Watcher, &actor.Terminated{Who: pid})
			}
			return
		}
		ref.mu.Lock()
		ref.watchers.Add(msg.Watcher)
		ref.mu.Unlock()

	case *actor.Unwatch:
		ref.mu.Lock()
		ref.watchers.Remove(msg.Watcher)
		ref.mu.Unlock()

	case *actor.Stop:
		term := &actor.Terminated{Who: pid}
		ref.mu.Lock()
		ref.watchers.ForEach(func(_ int, other *actor.PID) {
			if !other.Equal(ref.parent) {
				if r, ok := ref.actorSystem.ProcessRegistry.Get(other); ok {
					r.SendSystemMessage(other, term)
				}
			}
		})
		// Notify parent
		if ref.parent != nil {
			if r, ok := ref.actorSystem.ProcessRegistry.Get(ref.parent); ok {
				r.SendSystemMessage(ref.parent, term)
			}
		}
		ref.mu.Unlock()

		ref.finalizeStop(pid)
	default:
		r, _ := ref.actorSystem.ProcessRegistry.Get(ref.router)
		r.SendSystemMessage(pid, message)

	}
}

func (ref *routerProcess) finalizeStop(pid *actor.PID) {
	if atomic.SwapInt32(&ref.stopping, 1) == 1 {
		return
	}

	_ = ref.actorSystem.Root.StopFuture(ref.router).Wait()
	ref.actorSystem.ProcessRegistry.Remove(pid)
}

func (ref *routerProcess) Stop(pid *actor.PID) {
	ref.finalizeStop(pid)
	ref.SendSystemMessage(pid, &actor.Stop{})
}

func (ref *routerProcess) Poison(pid *actor.PID) {
	if atomic.SwapInt32(&ref.stopping, 1) == 1 {
		return
	}

	_ = ref.actorSystem.Root.PoisonFuture(ref.router).Wait()
	ref.actorSystem.ProcessRegistry.Remove(pid)
	ref.SendSystemMessage(pid, &actor.Stop{})
}
