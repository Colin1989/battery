package actor

import (
	"runtime"
	"sync/atomic"
)

type Mailbox interface {
	Start()
	Count() int
	Post(message MessageEnvelope)
}

// Invoker is the interface used by a mailbox to forward messages for processing
type Invoker interface {
	Invoke(interface{})
}

// MailboxProducer is a function which creates a new mailbox
type MailboxProducer func() Mailbox

const (
	idle int32 = iota
	running
)

type defaultMailbox struct {
	mb         queue
	dispatcher Dispatcher
	invoker    Invoker
	//middlewares
	schedulerStatus int32
	//suspended       int32
	messages int32
}

func (m *defaultMailbox) Start() {
}

func (m *defaultMailbox) Count() int {
	return int(atomic.LoadInt32(&m.messages))
}

func (m *defaultMailbox) Post(message MessageEnvelope) {
	m.mb.Push(message)
	atomic.AddInt32(&m.messages, 1)
	m.schedule()
}

func (m *defaultMailbox) schedule() {
	if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
		m.dispatcher.Schedule(m.processMessages)
	}
}

func (m *defaultMailbox) processMessages() {
	m.run()
	atomic.StoreInt32(&m.schedulerStatus, idle)
}

func (m *defaultMailbox) run() {
	var msg interface{}

	defer func() {
		if r := recover(); r != nil {
			//plog.Info("[ACTOR] Recovering", log.Object("actor", m.invoker), log.Object("reason", r), log.Stack())
			//m.invoker.EscalateFailure(r, msg)
		}
	}()

	i, t := 0, m.dispatcher.Throughput()
	for {
		if i > t {
			i = 0
			runtime.Gosched()
		}

		i++

		if msg = m.mb.Pop(); msg != nil {
			atomic.AddInt32(&m.messages, -1)
			m.invoker.Invoke(msg)
		} else {
			return
		}
	}
}
