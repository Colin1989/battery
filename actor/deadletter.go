package actor

type deadLetter struct {
	pid         *PID
	actorSystem *ActorSystem
}

// DeadLetterEvent
// 当有消息发送给一个不存在的 PID 时。 发布事件给所有的订阅者
type DeadLetterEvent struct {
	PID     *PID // nonexistent PID
	Message interface{}
	Sender  *PID // 发送者
}

func newDeadLetter(actorSystem *ActorSystem) *deadLetter {
	dl := &deadLetter{
		actorSystem: actorSystem,
	}
	dl.pid, _ = actorSystem.ProcessRegistry.Add(dl, "deadLetter")

	// subscribe DeadLetterEvent
	actorSystem.EventStream.Subscribe(func(msg interface{}) {
		dlEvent, ok := msg.(*DeadLetterEvent)
		if !ok {
			return
		}
		_ = dlEvent
		//m, ok := deadLetter.
	})

	return dl
}

func (dp *deadLetter) Send(pid *PID, message interface{}) {
	//TODO need add metrics

	dp.actorSystem.EventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: message,
		Sender:  nil,
	})
}

func (dp *deadLetter) Stop(pid *PID) {
	dp.Send(pid, stopMessage)
}
