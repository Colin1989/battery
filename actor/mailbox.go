package actor

import (
	"github.com/colin1989/battery/actor/queue/goring"
	"github.com/colin1989/battery/actor/queue/mpsc"
	"runtime"
	"sync/atomic"
)

type Mailbox interface {
	Start()
	Count() int
	PostUserMessage(message *MessageEnvelope)
	PostSystemMessage(message SystemMessage)
	RegisterHandlers(invoker Invoker, dispatcher Dispatcher)
}

// Invoker is the interface used by a mailbox to forward messages for processing
type Invoker interface {
	InvokeSystemMessage(message SystemMessage)
	InvokeUserMessage(envelope *MessageEnvelope)
	EscalateFailure(reason interface{}, message interface{})
}

// MailboxProducer is a function which creates a new mailbox
type MailboxProducer func() Mailbox

const (
	idle int32 = iota
	running
)

type defaultMailbox struct {
	userMailbox   queue[*MessageEnvelope]
	systemMailbox queue[SystemMessage]
	dispatcher    Dispatcher
	invoker       Invoker
	//middlewares
	schedulerStatus int32
	userMessages    int32
	sysMessages     int32
	suspended       int32
}

func newDefaultMailbox() Mailbox {
	mb := &defaultMailbox{
		userMailbox:     goring.New[*MessageEnvelope](10),
		systemMailbox:   mpsc.New[SystemMessage](),
		dispatcher:      nil,
		invoker:         nil,
		schedulerStatus: 0,
		userMessages:    0,
		sysMessages:     0,
		suspended:       0,
	}

	return mb
}

func (m *defaultMailbox) Start() {
}

func (m *defaultMailbox) Count() int {
	return int(atomic.LoadInt32(&m.userMessages))
}

func (m *defaultMailbox) PostUserMessage(message *MessageEnvelope) {
	m.userMailbox.Push(message)
	atomic.AddInt32(&m.userMessages, 1)
	m.schedule()
}

func (m *defaultMailbox) PostSystemMessage(message SystemMessage) {
	m.systemMailbox.Push(message)
	atomic.AddInt32(&m.sysMessages, 1)
	m.schedule()
}

func (m *defaultMailbox) RegisterHandlers(invoker Invoker, dispatcher Dispatcher) {
	m.invoker = invoker
	m.dispatcher = dispatcher
}

func (m *defaultMailbox) schedule() {
	if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
		m.dispatcher.Schedule(m.processMessages)
	}
}

func (m *defaultMailbox) processMessages() {
process:
	m.run()
	// set mailbox to idle
	atomic.StoreInt32(&m.schedulerStatus, idle)
	sys := atomic.LoadInt32(&m.sysMessages)
	user := atomic.LoadInt32(&m.userMessages)
	// check if there are still messages to process (sent after the message loop ended)
	if sys > 0 || (atomic.LoadInt32(&m.suspended) == 0 && user > 0) {
		// try setting the mailbox back to running
		if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
			//	fmt.Printf("looping %v %v %v\n", sys, user, m.suspended)
			goto process
		}
	}
}

func (m *defaultMailbox) run() {
	var envelope *MessageEnvelope
	var systemMessage SystemMessage
	var ok bool

	defer func() {
		if r := recover(); r != nil {
			//plog.Info("[ACTOR] Recovering", log.Object("actor", m.invoker), log.Object("reason", r), log.Stack())
			m.invoker.EscalateFailure(r, envelope)
		}
	}()

	i, t := 0, m.dispatcher.Throughput()
	for {
		if i > t {
			i = 0
			runtime.Gosched()
		}

		i++

		// keep processing system messages until queue is empty
		if systemMessage, ok = m.systemMailbox.Pop(); systemMessage != nil && ok {
			atomic.AddInt32(&m.sysMessages, -1)
			switch systemMessage {
			//case *SuspendMailbox:
			//	atomic.StoreInt32(&m.suspended, 1)
			//case *ResumeMailbox:
			//	atomic.StoreInt32(&m.suspended, 0)
			default:
				m.invoker.InvokeSystemMessage(systemMessage)
			}
			//for _, ms := range m.middlewares {
			//	ms.MessageReceived(envelope)
			//}
			continue
		}

		// didn't process a system message, so break until we are resumed
		if atomic.LoadInt32(&m.suspended) == 1 {
			return
		}

		if envelope, ok = m.userMailbox.Pop(); envelope != nil && ok {
			atomic.AddInt32(&m.userMessages, -1)
			m.invoker.InvokeUserMessage(envelope)
			//for _, ms := range m.middlewares {
			//	ms.MessageReceived(envelope)
			//}
		} else {
			return
		}
	}
}
