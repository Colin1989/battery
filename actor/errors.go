package actor

import "errors"

var (
	// ErrNameExists is the error used when an existing name is used for spawning an actor.
	ErrNameExists = errors.New("spawn: name exists")
	
	// ErrTimeout is the error used when a future times out before receiving a result.
	ErrTimeout = errors.New("future: timeout")

	// ErrDeadLetter is meaning you request to a unreachable PID.
	ErrDeadLetter = errors.New("future: dead letter")
)
