package actor

import (
	"github.com/lithammer/shortuuid/v4"
	"log/slog"
)

//goland:noinspection GoNameStartsWithPackageName
type ActorSystem struct {
	ProcessRegistry *ProcessRegistry
	Root            *RootContext
	EventStream     *EventStream
	DeadLetter      *deadLetter
	Config          *Config
	logger          *slog.Logger

	ID      string
	stopper chan struct{}
}

func NewActorSystem(opts ...ConfigOption) *ActorSystem {
	config := Configure(opts...)
	return NewActorSystemWithConfig(config)
}

func NewActorSystemWithConfig(config *Config) *ActorSystem {
	actorSystem := new(ActorSystem)
	actorSystem.Config = config
	actorSystem.ID = shortuuid.New()
	actorSystem.stopper = make(chan struct{}, 1)
	actorSystem.ProcessRegistry = NewProcessRegistry(actorSystem)
	actorSystem.Root = NewRootContext(actorSystem, EmptyMessageHeader)
	actorSystem.EventStream = NewEventStream()
	actorSystem.DeadLetter = newDeadLetter(actorSystem)
	actorSystem.logger = config.LoggerFactory(actorSystem)

	return actorSystem
}

func (as *ActorSystem) NewLocalPID(id string) *PID {
	return NewPID(as.ProcessRegistry.Address, id)
}

func (as *ActorSystem) Address() string {
	return as.ProcessRegistry.Address
}

func (as *ActorSystem) Shutdown() {
	as.ProcessRegistry.Remove(as.DeadLetter.pid)
	as.ProcessRegistry.shutdown()
	close(as.stopper)
}

func (as *ActorSystem) IsStopped() bool {
	select {
	case <-as.stopper:
		return true
	default:
		return false
	}
}
