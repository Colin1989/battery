package router

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/colin1989/battery/actor"
	"github.com/stretchr/testify/mock"
)

var nilPID *actor.PID

type mockContext struct {
	mock.Mock
}

func (m *mockContext) Parent() *actor.PID {
	args := m.Called()
	return args.Get(0).(*actor.PID)
}

func (m *mockContext) Self() *actor.PID {
	args := m.Called()
	return args.Get(0).(*actor.PID)
}

func (m *mockContext) Actor() actor.Actor {
	args := m.Called()
	return args.Get(0).(actor.Actor)
}

func (m *mockContext) ActorSystem() *actor.ActorSystem {
	args := m.Called()
	return args.Get(0).(*actor.ActorSystem)
}

func (m *mockContext) Logger() *slog.Logger {
	//TODO implement me
	panic("implement me")
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

func (m *mockContext) Children() []*actor.PID {
	args := m.Called()
	return args.Get(0).([]*actor.PID)
}

func (m *mockContext) Respond(envelope *actor.MessageEnvelope) {
	m.Called(envelope)
}

func (m *mockContext) Stash() {
	m.Called()
}

func (m *mockContext) Watch(pid *actor.PID) {
	m.Called(pid)
}

func (m *mockContext) Unwatch(pid *actor.PID) {
	m.Called(pid)
}

func (m *mockContext) Envelope() *actor.MessageEnvelope {
	args := m.Called()
	return args.Get(0).(*actor.MessageEnvelope)
}

func (m *mockContext) MessageHeader() actor.ReadonlyMessageHeader {
	args := m.Called()
	return args.Get(0).(actor.ReadonlyMessageHeader)
}

func (m *mockContext) Sender() *actor.PID {
	args := m.Called()
	return args.Get(0).(*actor.PID)
}

func (m *mockContext) Send(pid *actor.PID, envelope *actor.MessageEnvelope) {
	m.Called()
	p, _ := system.ProcessRegistry.Get(pid)
	p.SendUserMessage(pid, envelope)
}

func (m *mockContext) Request(pid *actor.PID, envelop *actor.MessageEnvelope) (*actor.MessageEnvelope, error) {
	args := m.Called(pid, envelop)
	return args.Get(0).(*actor.MessageEnvelope), args.Get(0).(error)
}

func (m *mockContext) Receive(envelope *actor.MessageEnvelope) {
	m.Called(envelope)
}

func (m *mockContext) Spawn(props *actor.Props) *actor.PID {
	args := m.Called(props)
	return args.Get(0).(*actor.PID)
}

func (m *mockContext) SpawnPrefix(props *actor.Props, prefix string) *actor.PID {
	args := m.Called(props, prefix)
	return args.Get(0).(*actor.PID)
}

func (m *mockContext) SpawnNamed(props *actor.Props, id string) (*actor.PID, error) {
	args := m.Called(props, id)
	return args.Get(0).(*actor.PID), args.Get(1).(error)
}

func (m *mockContext) Stop(pid *actor.PID) {
	m.Called(pid)
}

func (m *mockContext) StopFuture(pid *actor.PID) *actor.Future {
	args := m.Called(pid)
	return args.Get(0).(*actor.Future)
}

func (m *mockContext) Poison(pid *actor.PID) {
	m.Called(pid)
}

func (m *mockContext) PoisonFuture(pid *actor.PID) *actor.Future {
	args := m.Called(pid)
	return args.Get(0).(*actor.Future)
}

// mockProcess
type mockProcess struct {
	mock.Mock
}

func spawnMockProcess(name string) (*actor.PID, *mockProcess) {
	p := &mockProcess{}
	pid, ok := system.ProcessRegistry.Add(p, name)
	if !ok {
		panic(fmt.Errorf("did not spawn named process '%s'", name))
	}

	return pid, p
}

func (m *mockProcess) SendUserMessage(pid *actor.PID, envelope *actor.MessageEnvelope) {
	m.Called(pid, envelope)
}

func (m *mockProcess) SendSystemMessage(pid *actor.PID, message actor.SystemMessage) {
	m.Called(pid, message)
}

func removeMockProcess(pid *actor.PID) {
	system.ProcessRegistry.Remove(pid)
}

func (m *mockProcess) Stop(pid *actor.PID) {
	m.Called(pid)
}
