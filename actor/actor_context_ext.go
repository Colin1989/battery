package actor

import (
	"github.com/colin1989/battery/actor/ctxext"
	"github.com/emirpasic/gods/stacks/linkedliststack"
	"time"
)

type actorContextExtras struct {
	children            PIDSet
	receiveTimeoutTimer *time.Timer
	//rs                  *RestartStatistics
	stack      *linkedliststack.Stack
	watchers   PIDSet
	context    Context
	extensions *ctxext.ContextExtensions
}

func newActorContextExtras(context Context) *actorContextExtras {
	this := &actorContextExtras{
		context:    context,
		extensions: ctxext.NewContextExtensions(),
	}

	return this
}

//func (ctxExt *actorContextExtras) restartStats() *RestartStatistics {
//	// lazy initialize the child restart stats if this is the first time
//	// further mutations are handled within "restart"
//	if ctxExt.rs == nil {
//		ctxExt.rs = NewRestartStatistics()
//	}
//
//	return ctxExt.rs
//}

func (ctxExt *actorContextExtras) initReceiveTimeoutTimer(timer *time.Timer) {
	ctxExt.receiveTimeoutTimer = timer
}

func (ctxExt *actorContextExtras) resetReceiveTimeoutTimer(time time.Duration) {
	if ctxExt.receiveTimeoutTimer == nil {
		return
	}

	ctxExt.receiveTimeoutTimer.Reset(time)
}

func (ctxExt *actorContextExtras) stopReceiveTimeoutTimer() {
	if ctxExt.receiveTimeoutTimer == nil {
		return
	}

	ctxExt.receiveTimeoutTimer.Stop()
}

func (ctxExt *actorContextExtras) killReceiveTimeoutTimer() {
	if ctxExt.receiveTimeoutTimer == nil {
		return
	}

	ctxExt.receiveTimeoutTimer.Stop()
	ctxExt.receiveTimeoutTimer = nil
}

func (ctxExt *actorContextExtras) addChild(pid *PID) {
	ctxExt.children.Add(pid)
}

func (ctxExt *actorContextExtras) removeChild(pid *PID) {
	ctxExt.children.Remove(pid)
}

func (ctxExt *actorContextExtras) Children() []*PID {
	return ctxExt.children.Values()
}

func (ctxExt *actorContextExtras) watch(watcher *PID) {
	ctxExt.watchers.Add(watcher)
}

func (ctxExt *actorContextExtras) unwatch(watcher *PID) {
	ctxExt.watchers.Remove(watcher)
}

func (ctxExt *actorContextExtras) stash(message interface{}) {
	if ctxExt.stack == nil {
		ctxExt.stack = linkedliststack.New()
	}
	ctxExt.stack.Push(message)
}

func (ctxExt *actorContextExtras) popStash() (interface{}, bool) {
	if ctxExt.stack == nil || ctxExt.stack.Empty() {
		return nil, false
	}
	message, ok := ctxExt.stack.Pop()
	if !ok {
		return nil, false
	}
	return message, true
}
