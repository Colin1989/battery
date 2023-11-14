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

func (dp *deadLetter) SendUserMessage(pid *PID, message *MessageEnvelope) {
	_, msg, sender := UnwrapEnvelope(message)
	dp.actorSystem.EventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: msg,
		Sender:  sender,
	})
}

func (dp *deadLetter) SendSystemMessage(pid *PID, message SystemMessage) {
	//TODO need add metrics
	_, msg, _ := UnwrapEnvelope(message)
	dp.actorSystem.EventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: msg,
		Sender:  nil,
	})
}

func (dp *deadLetter) Stop(pid *PID) {
	dp.SendSystemMessage(pid, stopMessage)
}
