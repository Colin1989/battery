package message

type PendingMessage struct {
	Typ     Type        // message type
	Route   Route       // message route (push)
	Mid     uint        // response message id (response)
	Payload interface{} // payload
	Err     bool        // if its an error message
}

type BroadcastMessage struct {
	P []byte
}
