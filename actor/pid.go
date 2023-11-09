package actor

func NewPID(address, id string) *PID {
	return &PID{
		Address: address,
		ID:      id,
	}
}
