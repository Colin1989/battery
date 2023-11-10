package actor

import "time"

type Context interface {
	infoPart
	basePart
	messagePart
	senderPart
	spawnerPart
	stopperPart
}

type infoPart interface {
	// Parent returns the PID for the current actors parent
	Parent() *PID

	// Self returns the PID for the current actor
	Self() *PID

	// Actor returns the actor associated with this context
	Actor() Actor

	ActorSystem() *ActorSystem
}

type basePart interface {
	// ReceiveTimeout returns the current timeout
	ReceiveTimeout() time.Duration

	// Children returns a slice of the actors children
	Children() []*PID

	// Respond sends a response to the current `Sender`
	// If the Sender is nil, the actor will panic
	Respond(response interface{})
}

type messagePart interface {
	// Message returns the current message to be processed
	Message() interface{}

	// MessageHeader returns the meta information for the currently processed message
	MessageHeader() ReadonlyMessageHeader
}

type senderPart interface {
	// Sender returns the PID of actor that sent currently processed message
	Sender() *PID

	// Send sends a message to the given PID
	Send(pid *PID, message interface{})

	// SendEnvelope wrap the message into an envelope and send the envelope to the given PID.
	SendEnvelope(pid *PID, message interface{})
}

type spawnerPart interface {
	//// Spawn starts a new child actor based on props and named with a unique id
	//Spawn(props *Props) *PID
	//
	//// SpawnPrefix starts a new child actor based on props and named using a prefix followed by a unique id
	//SpawnPrefix(props *Props, prefix string) *PID
	//
	//// SpawnNamed starts a new child actor based on props and named using the specified name
	////
	//// ErrNameExists will be returned if id already exists
	////
	//// Please do not use name sharing same pattern with system actors, for example "YourPrefix$1", "Remote$1", "future$1"
	//SpawnNamed(props *Props, id string) (*PID, error)
}

type stopperPart interface {
	// Stop will stop actor immediately regardless of existing user messages in mailbox.
	Stop(pid *PID)

	// Poison will tell actor to stop after processing current user messages in mailbox.
	Poison(pid *PID)
}
