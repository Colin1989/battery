package mocks

import (
	"net"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockConnector struct {
	mock.Mock
}

func NewMockConnector() *mockConnector {
	return &mockConnector{}
}

func (m *mockConnector) GetNextMessage() (b []byte, err error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Get(1).(error)
}

func (m *mockConnector) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *mockConnector) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Get(0).(int), args.Get(1).(error)
}

func (m *mockConnector) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	if len(args) == 2 {
		return args.Get(0).(int), args.Get(1).(error)
	} else {
		return args.Get(0).(int), nil
	}
}

func (m *mockConnector) Close() error {
	args := m.Called()
	return args.Get(0).(error)
}

func (m *mockConnector) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *mockConnector) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Get(0).(error)
}

func (m *mockConnector) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Get(0).(error)
}

func (m *mockConnector) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Get(0).(error)
}
