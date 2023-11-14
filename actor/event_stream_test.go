package actor_test

import (
	"github.com/colin1989/battery/actor/actor"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestEvent struct {
	V int
}

func (e *TestEvent) EventMessage() {}

func TestEventStream_Subscribe(t *testing.T) {
	es := actor.NewEventStream()
	s := es.Subscribe(func(evt actor.EventMessage) {

	})
	assert.NotNil(t, s)
	assert.Equal(t, es.Length(), 1)
}

func TestEventStream_Unsubscribe(t *testing.T) {
	es := actor.NewEventStream()
	var e1, e2 int
	sub1 := es.Subscribe(func(evt actor.EventMessage) {
		e1++
	})
	sub2 := es.Subscribe(func(evt actor.EventMessage) {
		e2++
	})
	assert.Equal(t, es.Length(), 2)

	es.Unsubscribe(sub2)
	assert.Equal(t, es.Length(), 1)

	es.Publish(&TestEvent{V: 1})
	assert.Equal(t, e1, 1)

	es.Unsubscribe(sub1)
	assert.Equal(t, es.Length(), 0)

	es.Publish(&TestEvent{V: 1})
	assert.Equal(t, e1, 1)
	assert.Equal(t, e2, 0)
}

func TestEventStream_Publish(t *testing.T) {
	es := actor.NewEventStream()

	var v int
	es.Subscribe(func(evt actor.EventMessage) {
		m := evt.(*TestEvent)
		v = m.V
	})

	es.Publish(&TestEvent{V: 1})
	assert.Equal(t, v, 1)

	es.Publish(&TestEvent{V: 100})
	assert.Equal(t, v, 100)
}

func BenchmarkEventStream(b *testing.B) {
	es := actor.NewEventStream()
	subs := make([]*actor.Subscription, 10)
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			sub := es.Subscribe(func(evt actor.EventMessage) {
				if msg := evt.(*TestEvent); msg.V != i {
					b.Fatalf("expected i to be %d but its value is %d", i, msg.V)
				}
			})
			subs[j] = sub
		}

		es.Publish(&TestEvent{V: i})
		for j := range subs {
			es.Unsubscribe(subs[j])
			if subs[j].IsActive() {
				b.Fatal("subscription should not be active")
			}
		}
	}
}
