package actor

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type EventMessage interface {
	EventMessage()
}

// Handler defines a callback function that must be pass when subscribing.
type Handler func(evt EventMessage)

type EventStream struct {
	sync.RWMutex

	// slice containing our subscriptions
	subscriptions []*Subscription

	// Atomically maintained elements counter
	counter int32
}

func NewEventStream() *EventStream {
	es := &EventStream{
		subscriptions: make([]*Subscription, 0),
	}
	return es
}

// Subscribe the given handler to the EventStream
func (es *EventStream) Subscribe(handler Handler) *Subscription {
	sub := &Subscription{
		handler: handler,
		active:  1,
	}

	es.Lock()
	defer es.Unlock()
	sub.id = es.counter
	es.counter++
	es.subscriptions = append(es.subscriptions, sub)
	return sub
}

func (es *EventStream) Unsubscribe(sub *Subscription) {
	if sub == nil {
		return
	}

	if !sub.IsActive() {
		return
	}
	es.Lock()
	defer es.Unlock()

	// sub cannot deactivate twice
	if !sub.deActivate() {
		return
	}

	if es.counter == 0 {
		es.subscriptions = nil
		fmt.Printf("sub[%v] deActivate, but the counter is zero", sub.id)
		return
	}

	// swap
	l := es.counter - 1
	es.subscriptions[sub.id] = es.subscriptions[l]
	es.subscriptions[sub.id].id = sub.id
	es.subscriptions[l] = nil
	es.subscriptions = es.subscriptions[:l]
	es.counter--

	// clear
	if es.counter == 0 {
		es.subscriptions = nil
	}
}

func (es *EventStream) Publish(evt EventMessage) {
	subs := make([]*Subscription, 0, es.Length())
	es.RLock()
	for _, sub := range es.subscriptions {
		if !sub.IsActive() {
			continue
		}
		subs = append(subs, sub)
	}
	es.RUnlock()

	for _, sub := range subs {
		sub.handler(evt)
	}
}

func (es *EventStream) Length() int {
	es.RLock()
	defer es.RUnlock()
	return len(es.subscriptions)
}

// Subscription is returned from the Subscribe function.
//
// This value and can be passed to Unsubscribe when the observer is no longer interested in receiving messages
type Subscription struct {
	id      int32
	handler Handler
	active  uint32
}

func (s *Subscription) Activate() bool {
	return atomic.CompareAndSwapUint32(&s.active, 0, 1)
}

func (s *Subscription) deActivate() bool {
	return atomic.CompareAndSwapUint32(&s.active, 1, 0)
}

func (s *Subscription) IsActive() bool {
	return atomic.LoadUint32(&s.active) == 1
}
