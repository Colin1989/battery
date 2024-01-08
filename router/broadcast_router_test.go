package router

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/colin1989/battery/actor"
)

var system = actor.NewActorSystem()

func TestBroadcastRouterThreadSafe(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	props := actor.PropsFromFunc(func(c actor.Context) {
		switch c.Envelope().Message.(type) {
		case *AddRoutee:
			t.Logf("AddRoutee pid : %s", c.Self().String())
		case struct{}:
			t.Logf("struct{} pid : %s", c.Self().String())
		}
	})

	grp := system.Root.Spawn(NewBroadcastGroup())
	go func() {
		count := 100
		for i := 0; i < count; i++ {
			pid, _ := system.Root.SpawnNamed(props, strconv.Itoa(i))
			system.Root.Send(grp, AddRouteeEnvelope(pid))
			time.Sleep(10 * time.Millisecond)
		}
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	go func() {
		count := 100
		for c := 0; c < count; c++ {
			system.Root.Send(grp, actor.WrapEnvelope(struct{}{}))
			time.Sleep(10 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()
}
