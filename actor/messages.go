package actor

// A SystemMessage message is reserved for specific lifecycle messages used by the actor system
type SystemMessage interface {
	SystemMessage()
}

type Stop struct {
}

func (*Stop) SystemMessage() {}

var (
	stopMessage SystemMessage = &Stop{}
)
