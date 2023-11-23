package acceptor

import (
	"crypto/tls"
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/helper"
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
	{"test_1", "0.0.0.0:1234", []byte{0x01, 0x02}, []string{"./fixtures/server.crt", "./fixtures/server.key"}, nil},
	{"test_2", "127.0.0.1:1235", []byte{0x00}, []string{"./fixtures/server.crt", "./fixtures/server.key"}, nil},
	{"test_3", "0.0.0.0:1236", []byte{0x00}, []string{"wqodij"}, constant.ErrIncorrectNumberOfCertificates},
	{"test_4", "0.0.0.0:1237", []byte{0x00}, []string{"wqodij", "qwdo", "wod"}, constant.ErrIncorrectNumberOfCertificates},
	{"test_4", "0.0.0.0:1238", []byte{0x00}, []string{}, nil},
}

var (
	wsPID *actor.PID
)

func TestNewWsAcceptor(t *testing.T) {
	t.Parallel()
	for _, table := range wsAcceptorTables {
		t.Run(table.name, func(t *testing.T) {
			if table.panicErr != nil {
				assert.PanicsWithValue(t, table.panicErr, func() {
					NewWsAcceptor(table.addr, func(conn Connector) actor.Producer {
						return func() actor.Actor {
							return &tAgent{
								conn: conn,
							}
						}
					}, table.certs...)
				})
			} else {
				assert.NotPanics(t, func() {
					acceptorProducer := NewWsAcceptor(table.addr, func(conn Connector) actor.Producer {
						return func() actor.Actor {
							return &tAgent{
								conn: conn,
							}
						}
					}, table.certs...)
					wsPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), WSAcceptorName())
				})

				assert.NotNil(t, wsPID)
			}
		})
	}
}

//func TestWSAcceptor_GetAddr(t *testing.T) {
//	t.Parallel()
//	for _, tt := range wsAcceptorTables {
//		t.Run(tt.name, func(t *testing.T) {
//			w := NewWsAcceptor(tt.addr)
//			// will return empty string because acceptor is not listening
//			assert.Empty(t, w.GetAddr())
//			go w.ListenAndServe()
//			mustConnectToWS(t, tt.write, w, "ws")
//			assert.NotEmpty(t, w.GetAddr())
//		})
//	}
//}
//
//func TestWSAcceptor_GetConnChan(t *testing.T) {
//	t.Parallel()
//	for _, tt := range wsAcceptorTables {
//		t.Run(tt.name, func(t *testing.T) {
//			w := NewWsAcceptor(tt.addr)
//			got := w.GetConnChan()
//			assert.NotNil(t, got)
//		})
//	}
//}
//
//func TestWSAcceptor_ListenAndServe(t *testing.T) {
//	for _, tt := range wsAcceptorTables {
//		t.Run(tt.name, func(t *testing.T) {
//			w := NewWsAcceptor(tt.addr)
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, tt.write, w, "ws")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			assert.NotNil(t, conn)
//		})
//	}
//}
//
//func TestWSAcceptor_ListenAndServeTLS(t *testing.T) {
//	for _, tt := range wsAcceptorTables {
//		t.Run(tt.name, func(t *testing.T) {
//			w := NewWsAcceptor(tt.addr, "./fixtures/server.crt", "./fixtures/server.key")
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, tt.write, w, "wss")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			assert.NotNil(t, conn)
//		})
//	}
//}
//
//func TestWSAcceptor_Stop(t *testing.T) {
//	for _, tt := range wsAcceptorTables {
//		t.Run(tt.name, func(t *testing.T) {
//			w := NewWsAcceptor(tt.addr)
//			go w.ListenAndServe()
//			mustConnectToWS(t, tt.write, w, "ws")
//			w.Stop()
//			addr := fmt.Sprintf("ws://%s", w.GetAddr())
//			_, _, err := websocket.DefaultDialer.Dial(addr, nil)
//			assert.Error(t, err)
//		})
//	}
//}
//
//func TestWSConnRead(t *testing.T) {
//	for _, table := range wsAcceptorTables {
//		t.Run(table.name, func(t *testing.T) {
//			w := NewWsAcceptor(table.addr)
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, table.write, w, "ws")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			b := make([]byte, len(table.write))
//			n, err := conn.Read(b)
//			assert.NoError(t, err)
//			assert.Equal(t, len(table.write), n)
//			assert.Equal(t, table.write, b)
//		})
//	}
//}
//
//func TestWSConnWrite(t *testing.T) {
//	for _, table := range wsAcceptorTables {
//		t.Run(table.name, func(t *testing.T) {
//			w := NewWsAcceptor(table.addr)
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, table.write, w, "ws")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			n, err := conn.Write(table.write)
//			assert.NoError(t, err)
//			assert.Equal(t, n, len(table.write))
//		})
//	}
//}
//
//func TestWSConnLocalAddr(t *testing.T) {
//	for _, table := range wsAcceptorTables {
//		t.Run(table.name, func(t *testing.T) {
//			w := NewWsAcceptor(table.addr)
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, table.write, w, "ws")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			assert.NotEmpty(t, conn.LocalAddr().String())
//		})
//	}
//}
//
//func TestWSConnRemoteAddr(t *testing.T) {
//	for _, table := range wsAcceptorTables {
//		t.Run(table.name, func(t *testing.T) {
//			w := NewWsAcceptor(table.addr)
//			defer w.Stop()
//			c := w.GetConnChan()
//			go w.ListenAndServe()
//			mustConnectToWS(t, table.write, w, "ws")
//			conn := helper.ShouldEventuallyReceive(t, c, 100*time.Millisecond).(*WSConn)
//			defer conn.Close()
//			assert.NotEmpty(t, conn.RemoteAddr().String())
//		})
//	}
//}
//
//func TestWSConnSetDeadline(t *testing.T) {
//	for _, table := range wsAcceptorTables {
//		t.Run(table.name, func(t *testing.T) {
//			w := NewWsAcceptor(table.addr)
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
//
//func TestWSNextMessage(t *testing.T) {
//	wsTables := []struct {
//		name string
//		data []byte
//		err  error
//	}{
//		{"invalid_header", []byte{0x00, 0x00, 0x00, 0x00}, packet.ErrWrongPomeloPacketType},
//		{"valid_message", []byte{0x02, 0x00, 0x00, 0x01, 0x00}, nil},
//		{"invalid_message", []byte{0x02, 0x00, 0x00, 0x02, 0x00}, constant.ErrReceivedMsgSmallerThanExpected},
//		{"invalid_header", []byte{0x02, 0x00}, packet.ErrInvalidPomeloHeader},
//	}
//	for _, table := range wsTables {
//		t.Run(table.name, func(t *testing.T) {
//			ws := NewWsAcceptor("0.0.0.0:0")
//			c := ws.GetConnChan()
//			defer ws.Stop()
//			go ws.ListenAndServe()
//
//			var conn *websocket.Conn
//			var err error
//			helper.ShouldEventuallyReturn(t, func() error {
//				addr := fmt.Sprintf("ws://%s", ws.GetAddr())
//				dialer := websocket.DefaultDialer
//				conn, _, err = dialer.Dial(addr, nil)
//				if err != nil {
//					return err
//				}
//				return nil
//			}, nil, time.Millisecond*30, time.Millisecond*100)
//			wsConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*WSConn)
//			defer wsConn.Close()
//			err = conn.WriteMessage(websocket.BinaryMessage, table.data)
//			assert.NoError(t, err)
//			message, err := wsConn.GetNextMessage()
//			if table.err != nil {
//				assert.EqualError(t, err, table.err.Error())
//			} else {
//				assert.NoError(t, err)
//				assert.Equal(t, table.data, message)
//			}
//		})
//	}
//}
//
//func TestWSGetNextMessageSequentially(t *testing.T) {
//	testTables := []struct {
//		write   []byte
//		receive []byte
//		err     error
//	}{
//		{
//			write:   []byte{0x01, 0x00, 0x00, 0x02, 0x01, 0x01},
//			receive: []byte{0x01, 0x00, 0x00, 0x02, 0x01, 0x01},
//		},
//		{
//			write:   []byte{0x02, 0x00, 0x00, 0x02, 0x05, 0x04},
//			receive: []byte{0x02, 0x00, 0x00, 0x02, 0x05, 0x04},
//		},
//		{
//			write: []byte{0x02, 0x00, 0x00, 0x04},
//			err:   constant.ErrReceivedMsgSmallerThanExpected,
//		},
//		{
//			write: []byte{0x00, 0x00, 0x00},
//			err:   packet.ErrInvalidPomeloHeader,
//		},
//		{
//			write: []byte{0x00, 0x00, 0x00, 0x00},
//			err:   packet.ErrWrongPomeloPacketType,
//		},
//	}
//
//	ws := NewWsAcceptor("0.0.0.0:0")
//	defer ws.Stop()
//	c := ws.GetConnChan()
//	go ws.ListenAndServe()
//
//	var conn *websocket.Conn
//	var msgBytes []byte
//	var err error
//	helper.ShouldEventuallyReturn(t, func() error {
//		addr := fmt.Sprintf("ws://%s", ws.GetAddr())
//		dialer := websocket.DefaultDialer
//		conn, _, err = dialer.Dial(addr, nil)
//		return err
//	}, nil, time.Millisecond*30, time.Millisecond*100)
//
//	wsConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*WSConn)
//	defer wsConn.Close()
//
//	for _, table := range testTables {
//		err = conn.WriteMessage(websocket.BinaryMessage, table.write)
//		assert.NoError(t, err)
//	}
//
//	for _, table := range testTables {
//		msgBytes, err = wsConn.GetNextMessage()
//		if table.err != nil {
//			assert.EqualError(t, err, table.err.Error())
//		} else {
//			assert.NoError(t, err)
//			assert.Equal(t, table.write, msgBytes)
//		}
//	}
//
//	//msg1 := []byte{0x01, 0x00, 0x00, 0x02, 0x01, 0x01}
//	//msg2 := []byte{0x02, 0x00, 0x00, 0x02, 0x05, 0x04}
//	//err = conn.WriteMessage(websocket.BinaryMessage, msg1)
//	//assert.NoError(t, err)
//	//err = conn.WriteMessage(websocket.BinaryMessage, msg2)
//	//assert.NoError(t, err)
//	//
//	//msgBytes, err = wsConn.GetNextMessage()
//	//assert.NoError(t, err)
//	//assert.Equal(t, msg1, msgBytes)
//	//
//	//msgBytes, err = wsConn.GetNextMessage()
//	//assert.NoError(t, err)
//	//assert.Equal(t, msg2, msgBytes)
//}

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
