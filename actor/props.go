package actor

import "errors"

// Default values.
var (
	defaultDispatcher      = NewDefaultDispatcher(300)
	defaultMailboxProducer = newDefaultMailbox
	defaultSpawner         = func(actorSystem *ActorSystem, id string, props *Props, parentContext SpawnerContext) (*PID, error) {
		ctx := newActorContext(actorSystem, props, parentContext.Self())
		mb := props.produceMailbox()

		dp := props.getDispatcher()
		proc := NewActorProcess(mb)
		pid, absent := actorSystem.ProcessRegistry.Add(proc, id)
		if !absent {
			return pid, errors.New("spawn: name exists")
		}
		ctx.self = pid

		initialize(props, ctx)

		mb.RegisterHandlers(ctx, dp)
		mb.PostSystemMessage(makeMessage[Started]())
		mb.Start()

		return pid, nil
	}
)

func initialize(props *Props, ctx *actorContext) {
	if props.onInit == nil {
		return
	}

	for _, init := range props.onInit {
		init(ctx)
	}
}

type Props struct {
	spawner         SpawnFunc
	producer        Producer
	mailboxProducer MailboxProducer
	dispatcher      Dispatcher
	//receiverMiddleware      []ReceiverMiddleware
	//senderMiddleware        []SenderMiddleware
	//spawnMiddleware         []SpawnMiddleware
	receiverMiddlewareChain ReceiverFunc
	senderMiddlewareChain   SenderFunc
	spawnMiddlewareChain    SpawnFunc
	onInit                  []func(ctx Context)
}

func (props *Props) getSpawner() SpawnFunc {
	if props.spawner == nil {
		return defaultSpawner
	}

	return props.spawner
}

func (props *Props) getDispatcher() Dispatcher {
	if props.dispatcher == nil {
		return defaultDispatcher
	}

	return props.dispatcher
}

func (props *Props) produceMailbox() Mailbox {
	if props.mailboxProducer == nil {
		return defaultMailboxProducer()
	}

	return props.mailboxProducer()
}

func (props *Props) spawn(actorSystem *ActorSystem, name string, parentContext SpawnerContext) (*PID, error) {
	return props.getSpawner()(actorSystem, name, props, parentContext)
}

func (props *Props) Configure(opts ...PropsOption) *Props {
	for _, opt := range opts {
		opt(props)
	}

	return props
}
