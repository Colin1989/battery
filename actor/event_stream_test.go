package actor_test

import (
	"github.com/colin1989/battery/actor/actor"
	"testing"
)

func TestEventStream(t *testing.T) {
	es := actor.NewEventStream()
	sub := es.Subscribe(func(msg interface{}) {
		if _, ok := msg.(int); ok {
			t.Logf("Subscribe :%+v", msg)
		} else {
			t.Logf("Subscribe :%+v unspport type : %T", msg, msg)
		}
	})
	t.Logf("es.Publish(1)")
	es.Publish(1)
	t.Logf("es.Publish(\"2\")")
	es.Publish("2")
	t.Logf("event stream length : %+v", es.Length())
	es.Unsubscribe(sub)
	t.Logf("es.Unsubscribe(sub)")
	t.Logf("event stream length : %+v", es.Length())
}
