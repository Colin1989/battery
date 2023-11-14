package actor

import (
	"sync/atomic"
	"unsafe"
)

func NewPID(address, id string) *PID {
	return &PID{
		Address: address,
		ID:      id,
	}
}

//goland:noinspection GoReceiverNames
func (pid *PID) ref(actorSystem *ActorSystem) Process {
	p := (*Process)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&pid.p))))
	if p != nil {
		if l, ok := (*p).(*ActorProcess); ok && atomic.LoadInt32(&l.dead) == 1 {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pid.p)), nil)
		} else {
			return *p
		}
	}

	ref, exists := actorSystem.ProcessRegistry.Get(pid)
	if exists {
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pid.p)), unsafe.Pointer(&ref))
	}

	return ref
}

// sendUserMessage sends a messages asynchronously to the PID.
//
//goland:noinspection GoReceiverNames
func (pid *PID) sendUserMessage(actorSystem *ActorSystem, envelope *MessageEnvelope) {
	pid.ref(actorSystem).SendUserMessage(pid, envelope)
}

//goland:noinspection GoReceiverNames.
func (pid *PID) sendSystemMessage(actorSystem *ActorSystem, message SystemMessage) {
	pid.ref(actorSystem).SendSystemMessage(pid, message)
}
