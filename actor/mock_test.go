package actor

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"time"
)

var (
	system      = NewActorSystem()
	rootContext = system.Root
)

type mockContext struct {
	mock.Mock
}

func (m *mockContext) Parent() *PID {
	args := m.Called()
	return args.Get(0).(*PID)
}

func (m *mockContext) Self() *PID {
	args := m.Called()
	return args.Get(0).(*PID)
}

func (m *mockContext) Actor() Actor {
	args := m.Called()
	return args.Get(0).(Actor)
}

func (m *mockContext) ActorSystem() *ActorSystem {
	args := m.Called()
	return args.Get(0).(*ActorSystem)
}

func (m *mockContext) ReceiveTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockContext) SetReceiveTimeout(d time.Duration) {
	m.Called(d)
}

func (m *mockContext) CancelReceiveTimeout() {
	m.Called()
}

func (m *mockContext) Children() []*PID {
	args := m.Called()
	return args.Get(0).([]*PID)
}

func (m *mockContext) Respond(envelope *MessageEnvelope) {
	m.Called(envelope)
}

func (m *mockContext) Envelope() *MessageEnvelope {
	args := m.Called()
	return args.Get(0).(*MessageEnvelope)
}

func (m *mockContext) MessageHeader() ReadonlyMessageHeader {
	args := m.Called()
	return args.Get(0).(ReadonlyMessageHeader)
}

func (m *mockContext) Sender() *PID {
	args := m.Called()
	return args.Get(0).(*PID)
}

func (m *mockContext) Send(_ *PID, _ *MessageEnvelope) {
	m.Called()
}

func (m *mockContext) Request(pid *PID, message interface{}) (*MessageEnvelope, error) {
	args := m.Called(pid, message)
	return args.Get(0).(*MessageEnvelope), args.Get(0).(error)
}

func (m *mockContext) Receive(envelope *MessageEnvelope) {
	m.Called(envelope)
}

func (m *mockContext) Spawn(props *Props) *PID {
	args := m.Called(props)
	return args.Get(0).(*PID)
}

func (m *mockContext) SpawnPrefix(props *Props, prefix string) *PID {
	args := m.Called(props, prefix)
	return args.Get(0).(*PID)
}

func (m *mockContext) SpawnNamed(props *Props, id string) (*PID, error) {
	args := m.Called(props, id)
	return args.Get(0).(*PID), args.Get(0).(error)
}

func (m *mockContext) Stop(pid *PID) {
	m.Called(pid)
}

func (m *mockContext) Poison(pid *PID) {
	m.Called(pid)
}

type mockProcess struct {
	mock.Mock
}

func spawnMockProcess(name string) (*PID, *mockProcess) {
	p := &mockProcess{}

	pid, ok := system.ProcessRegistry.Add(p, name)
	if !ok {
		panic(fmt.Errorf("did not spawn named process '%vids'", name))
	}

	return pid, p
}

func removeMockProcess(pid *PID) {
	system.ProcessRegistry.Remove(pid)
}

func (m *mockProcess) SendUserMessage(pid *PID, envelope *MessageEnvelope) {
	m.Called(pid, envelope)
}

func (m *mockProcess) SendSystemMessage(pid *PID, message SystemMessage) {
	m.Called(pid, message)
}

func (m *mockProcess) Stop(pid *PID) {
	m.Called(pid)
}
