package actor

import "github.com/lithammer/shortuuid/v4"

//goland:noinspection GoNameStartsWithPackageName
type ActorSystem struct {
	ProcessRegistry *ProcessRegistry
	Root            *RootContext
	DeadLetter      *deadLetter
	EventStream     *EventStream
	Config          *Config

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
	actorSystem.DeadLetter = newDeadLetter(actorSystem)
	actorSystem.EventStream = NewEventStream()

	return actorSystem
}
