package acceptor

import (
	"crypto/tls"
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/helper"
	"github.com/colin1989/battery/net/packet"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var wsAcceptorTables = []struct {
	name     string
	addr     string
	write    []byte
	certs    []string
	panicErr error
}{
	// TODO change to allocatable ports
	{"test_1", "0.0.0.0:0", []byte{0x01, 0x02}, []string{"./fixtures/server.crt", "./fixtures/server.key"}, nil},
	{"test_2", "127.0.0.1:0", []byte{0x00}, []string{"./fixtures/server.crt", "./fixtures/server.key"}, nil},
	{"test_3", "0.0.0.0:0", []byte{0x00}, []string{"wqodij"}, constant.ErrIncorrectNumberOfCertificates},
	{"test_4", "0.0.0.0:0", []byte{0x00}, []string{"wqodij", "qwdo", "wod"}, constant.ErrIncorrectNumberOfCertificates},
	{"test_5", "0.0.0.0:0", []byte{0x00}, []string{}, nil},
}

var (
	wsActors = make([]*actor.PID, 0)
)

func TestNewWsAcceptor(t *testing.T) {
	t.Parallel()
	for _, table := range wsAcceptorTables {
		t.Run(table.name, func(t *testing.T) {
			var wsPID *actor.PID
			if table.panicErr != nil {
				assert.PanicsWithValue(t, table.panicErr, func() {
					wsPID = system.Root.SpawnPrefix(actor.PropsFromProducer(func() actor.Actor {
						return NewWSAcceptor(table.addr, table.certs...)
					}), constant.WSAcceptor)
				})
			} else {
				assert.NotPanics(t, func() {
					wsPID = system.Root.SpawnPrefix(actor.PropsFromProducer(func() actor.Actor {
						return NewWSAcceptor(table.addr, table.certs...)
					}), constant.WSAcceptor)
				})

				assert.NotNil(t, wsPID)
			}
		})
	}
}

func TestWSAcceptor_GetAddr(t *testing.T) {
	t.Parallel()
	for _, table := range wsAcceptorTables {
		t.Run(table.name, func(t *testing.T) {
			var w *WSAcceptor
			var wsPID *actor.PID
			props := actor.PropsFromProducer(func() actor.Actor {
				w = NewWSAcceptor(table.addr)
				return w
			})
			wsPID = system.Root.SpawnPrefix(props, constant.WSAcceptor)
			wsActors = append(wsActors, wsPID)
			// will return empty string because acceptor is not listening
			assert.Empty(t, w.GetAddr())
			time.Sleep(time.Millisecond * 500)
			mustConnectToWS(t, table.write, w.GetAddr(), "ws")
			assert.NotEmpty(t, w.GetAddr())
		})
	}

	time.Sleep(time.Second)
	assert.NotNil(t, agentPid)
	respCount, err := system.Root.Request(agentManagerPID, &ReqChildCount{})
	assert.NoError(t, err)
	assert.Equal(t, len(wsAcceptorTables), respCount.Message.(int))
	system.Root.Poison(agentManagerPID)

	for _, wsActor := range wsActors {
		system.Root.Poison(wsActor)
	}
	system.Shutdown()
}

func TestWSNextMessage(t *testing.T) {
	wsTables := []struct {
		name string
		data []byte
		err  error
	}{
		{"invalid_header", []byte{0x00, 0x00, 0x00, 0x00}, packet.ErrWrongPomeloPacketType},
		{"valid_message", []byte{0x02, 0x00, 0x00, 0x01, 0x00}, nil},
		{"invalid_message", []byte{0x02, 0x00, 0x00, 0x02, 0x00}, constant.ErrReceivedMsgSmallerThanExpected},
		{"invalid_header", []byte{0x02, 0x00}, packet.ErrInvalidPomeloHeader},
	}
	t.Parallel()
	for _, table := range wsTables {
		t.Run(table.name, func(t *testing.T) {
			var w *WSAcceptor
			var wsPID *actor.PID
			props := actor.PropsFromProducer(func() actor.Actor {
				w = NewWSAcceptor("0.0.0.0:0")
				return w
			})
			wsPID = system.Root.SpawnPrefix(props, constant.WSAcceptor)
			wsActors = append(wsActors, wsPID)
			// will return empty string because acceptor is not listening
			assert.Empty(t, w.GetAddr())
			time.Sleep(time.Millisecond * 500)
			mustConnectToWS(t, table.data, w.GetAddr(), "ws")
			assert.NotEmpty(t, w.GetAddr())
			time.Sleep(time.Millisecond * 500)

			request, err := system.Root.Request(agentPid, &requestMsg{})
			assert.NoError(t, err)
			if table.err != nil {
				assert.Equal(t, table.err, request.Message.(error))
			} else {
				assert.Equal(t, len(table.data), len(request.Message.([]byte)))
				assert.Equal(t, table.data, request.Message.([]byte))
			}
		})
	}
}

func TestWSConnLocalAddrAndRemoteAddr(t *testing.T) {
	for _, table := range wsAcceptorTables {
		t.Run(table.name, func(t *testing.T) {
			var w *WSAcceptor
			var wsPID *actor.PID
			props := actor.PropsFromProducer(func() actor.Actor {
				w = NewWSAcceptor("0.0.0.0:0")
				return w
			})
			wsPID = system.Root.SpawnPrefix(props, constant.WSAcceptor)
			wsActors = append(wsActors, wsPID)
			// will return empty string because acceptor is not listening
			assert.Empty(t, w.GetAddr())
			time.Sleep(time.Millisecond * 500)
			mustConnectToWS(t, table.write, w.GetAddr(), "ws")
			assert.NotEmpty(t, w.GetAddr())
			time.Sleep(time.Millisecond * 500)

			request, err := system.Root.Request(agentPid, &localAddr{})
			assert.NoError(t, err)
			assert.NotNil(t, request.Message)
			assert.NotEmpty(t, request.Message)

			request, err = system.Root.Request(agentPid, &remoteAddr{})
			assert.NoError(t, err)
			assert.NotNil(t, request.Message)
			assert.NotEmpty(t, request.Message)
		})
	}
}

//func TestWSConnSetDeadline(t *testing.T) {
//	for _, table := range wsAcceptorTables {
//		t.Run(table.name, func(t *testing.T) {
//			w := NewWSAcceptor(table.addr)
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, table.write, w, "ws")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			conn.SetDeadline(time.Now().Add(5 * time.Second))
//			time.Sleep(10 * time.Second)
//			_, err := conn.Write(table.write)
//			assert.Error(t, err)
//		})
//	}
//}

func TestWSGetNextMessageSequentially(t *testing.T) {
	testTables := []struct {
		write   []byte
		receive []byte
		err     error
	}{
		{
			write:   []byte{0x01, 0x00, 0x00, 0x02, 0x01, 0x01},
			receive: []byte{0x01, 0x00, 0x00, 0x02, 0x01, 0x01},
		},
		{
			write:   []byte{0x02, 0x00, 0x00, 0x02, 0x05, 0x04},
			receive: []byte{0x02, 0x00, 0x00, 0x02, 0x05, 0x04},
		},
		{
			write: []byte{0x02, 0x00, 0x00, 0x04},
			err:   constant.ErrReceivedMsgSmallerThanExpected,
		},
		{
			write: []byte{0x00, 0x00, 0x00},
			err:   packet.ErrInvalidPomeloHeader,
		},
		{
			write: []byte{0x00, 0x00, 0x00, 0x00},
			err:   packet.ErrWrongPomeloPacketType,
		},
	}

	var ws *WSAcceptor
	var wsPID *actor.PID
	props := actor.PropsFromProducer(func() actor.Actor {
		ws = NewWSAcceptor("0.0.0.0:0")
		return ws
	})
	wsPID = system.Root.SpawnPrefix(props, constant.WSAcceptor)
	wsActors = append(wsActors, wsPID)

	var conn *websocket.Conn
	var err error
	helper.ShouldEventuallyReturn(t, func() error {
		addr := fmt.Sprintf("ws://%s", ws.GetAddr())
		dialer := websocket.DefaultDialer
		conn, _, err = dialer.Dial(addr, nil)
		return err
	}, nil, time.Millisecond*30, time.Millisecond*100)

	for _, table := range testTables {
		err = conn.WriteMessage(websocket.BinaryMessage, table.write)
		assert.NoError(t, err)
	}

	time.Sleep(time.Millisecond * 100)
	for _, table := range testTables {
		request, err := system.Root.Request(agentPid, &requestMsg{})
		assert.NoError(t, err)
		if table.err != nil {
			assert.EqualError(t, request.Message.(error), table.err.Error())
		} else {
			assert.Equal(t, request.Message.([]byte), table.write)
		}
	}
}

func mustConnectToWS(t *testing.T, write []byte, addr, protocol string) {
	t.Helper()

	helper.ShouldEventuallyReturn(t, func() error {
		addr := fmt.Sprintf("%s://%s", protocol, addr)
		dialer := websocket.DefaultDialer
		conn, _, err := dialer.Dial(addr, nil)
		if err != nil {
			return err
		}
		defer conn.Close()
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		conn.WriteMessage(websocket.BinaryMessage, write)
		return err
	}, nil, 30*time.Millisecond, 100*time.Millisecond)
}
