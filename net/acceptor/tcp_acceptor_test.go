package acceptor

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/helper"
	"github.com/colin1989/battery/net/packet"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

type tAgent struct {
	conn    Connector
	message []interface{}
}

type requestMsg struct{}

func (ta *tAgent) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		agentPid = ctx.Self()
		fmt.Printf("tAgent started \n")
		go ta.read(ctx)
	case *actor.Stopped:
		err := ta.conn.Close()
		fmt.Printf("tAgent stopped err [%v]  \n", err)
	case *requestMsg:
		var resp interface{}
		if len(ta.message) > 0 {
			resp = ta.message[0]
			ta.message = ta.message[1:]
		}
		ctx.Respond(actor.WrapEnvelop(resp))
	default:
		ta.message = append(ta.message, msg)
	}
}

func (ta *tAgent) read(ctx actor.Context) {
	for {
		message, err := ta.conn.GetNextMessage()
		if err != nil {
			fmt.Printf("pid[%v] conn receive err[%v]", ctx.Self().String(), err)
			ctx.Send(ctx.Self(), actor.WrapEnvelop(err))
			//ctx.Poison(ctx.Self())
			return
		}
		ctx.Send(ctx.Self(), actor.WrapEnvelop(message))
	}
}

var tcpAcceptorTables = []struct {
	name     string
	addr     string
	certs    []string
	panicErr error
}{
	{"test_1", "0.0.0.0:1234", []string{"./fixtures/server.crt", "./fixtures/server.key"}, nil},
	{"test_2", "0.0.0.0:1235", []string{}, nil},
	{"test_3", "127.0.0.1:1236", []string{"wqd"}, constant.ErrIncorrectNumberOfCertificates},
	{"test_4", "127.0.0.1:1237", []string{"wqd", "wqdqwd"}, fmt.Errorf("%w: %v", constant.ErrInvalidCertificates, "open wqd: The system cannot find the file specified.")},
	{"test_5", "127.0.0.1:1238", []string{"wqd", "wqdqwd", "wqdqdqwd"}, constant.ErrIncorrectNumberOfCertificates},
}

var (
	system   = actor.NewActorSystem()
	tcpPID   *actor.PID
	agentPid *actor.PID
)

func TestNewTCPAcceptor(t *testing.T) {
	for _, table := range tcpAcceptorTables {
		t.Run(table.name, func(t *testing.T) {
			if table.panicErr != nil {
				assert.PanicsWithError(t, table.panicErr.Error(), func() {
					acceptorProducer := NewTCPAcceptor(table.addr, func(conn Connector) actor.Producer {
						return func() actor.Actor {
							return &tAgent{
								conn: conn,
							}
						}
					}, table.certs...)
					system.Root.Spawn(actor.PropsFromProducer(acceptorProducer))
				})
			} else {
				assert.NotPanics(t, func() {
					acceptorProducer := NewTCPAcceptor(table.addr, func(conn Connector) actor.Producer {
						return func() actor.Actor {
							return &tAgent{
								conn: conn,
							}
						}
					}, table.certs...)
					tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())
				})
				assert.NotNil(t, tcpPID)

				var conn net.Conn
				var err error
				// should be able to connect within 100 milliseconds
				helper.ShouldEventuallyReturn(t, func() error {
					conn, err = net.Dial("tcp", table.addr)
					defer conn.Close()
					return err
				}, nil, time.Millisecond*10, time.Millisecond*100)
				//conn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100)
				time.Sleep(time.Millisecond * 10)
				assert.NotNil(t, agentPid)
			}
		})
	}
}

func TestGetNextMessage(t *testing.T) {
	tables := []struct {
		name string
		addr string
		data []byte
		err  error
	}{
		{"invalid_header", "0.0.0.0:2234", []byte{0x00, 0x00, 0x00, 0x00}, packet.ErrWrongPomeloPacketType},
		{"valid_message", "0.0.0.0:2235", []byte{0x02, 0x00, 0x00, 0x01, 0x00}, nil},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			acceptorProducer := NewTCPAcceptor(table.addr, func(conn Connector) actor.Producer {
				return func() actor.Actor {
					return &tAgent{
						conn: conn,
					}
				}
			})
			tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())

			var conn net.Conn
			var err error
			helper.ShouldEventuallyReturn(t, func() error {
				conn, err = net.Dial("tcp", table.addr)
				return err
			}, nil, time.Millisecond*30, time.Millisecond*100)
			//tcpConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*TCPConn)

			write, err := conn.Write(table.data)
			assert.NoError(t, err)
			assert.Equal(t, len(table.data), write)

			time.Sleep(time.Second * 1)
			request, err := system.Root.Request(agentPid, &requestMsg{})
			assert.NoError(t, err)
			if table.err != nil {
				assert.EqualError(t, table.err, request.Message.(error).Error())
			} else {
				assert.Equal(t, table.data, request.Message)
			}
		})
	}
}

func TestGetNextMessageTwoMessagesInBuffer(t *testing.T) {
	addr := "0.0.0.0:3234"
	acceptorProducer := NewTCPAcceptor(addr, func(conn Connector) actor.Producer {
		return func() actor.Actor {
			return &tAgent{
				conn: conn,
				//err:     table.err,
				//message: table.data,
			}
		}
	})
	tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())

	var conn net.Conn
	var err error
	helper.ShouldEventuallyReturn(t, func() error {
		conn, err = net.Dial("tcp", addr)
		return err
	}, nil, time.Millisecond*30, time.Millisecond*100)

	//tcpConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*TCPConn)
	msg1 := []byte{0x01, 0x00, 0x00, 0x01, 0x02}
	msg2 := []byte{0x02, 0x00, 0x00, 0x02, 0x01, 0x01}
	buffer := append(msg1, msg2...)
	write, err := conn.Write(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(buffer), write)

	time.Sleep(time.Second * 1)
	request, err := system.Root.Request(agentPid, &requestMsg{})
	assert.NoError(t, err)
	assert.Equal(t, msg1, request.Message)
	request, err = system.Root.Request(agentPid, &requestMsg{})
	assert.NoError(t, err)
	assert.Equal(t, msg2, request.Message)
}

func TestGetNextMessageEOF(t *testing.T) {
	addr := "0.0.0.0:4234"
	acceptorProducer := NewTCPAcceptor(addr, func(conn Connector) actor.Producer {
		return func() actor.Actor {
			return &tAgent{
				conn: conn,
				//err:     table.err,
				//message: table.data,
			}
		}
	})
	tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())

	var conn net.Conn
	var err error
	helper.ShouldEventuallyReturn(t, func() error {
		conn, err = net.Dial("tcp", addr)
		return err
	}, nil, time.Millisecond*30, time.Millisecond*100)

	//tcpConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*TCPConn)
	buffer := []byte{0x02, 0x00, 0x00, 0x02, 0x01}
	write, err := conn.Write(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(buffer), write)

	go func() {
		time.Sleep(time.Millisecond * 100)
		conn.Close()
	}()

	time.Sleep(time.Millisecond * 200)
	request, err := system.Root.Request(agentPid, &requestMsg{})
	assert.NoError(t, err)
	assert.EqualError(t, constant.ErrReceivedMsgSmallerThanExpected, request.Message.(error).Error())
}

func TestGetNextMessageEmptyEOF(t *testing.T) {
	addr := "0.0.0.0:5234"
	acceptorProducer := NewTCPAcceptor(addr, func(conn Connector) actor.Producer {
		return func() actor.Actor {
			return &tAgent{
				conn: conn,
				//err:     table.err,
				//message: table.data,
			}
		}
	})
	tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())

	var conn net.Conn
	var err error
	helper.ShouldEventuallyReturn(t, func() error {
		conn, err = net.Dial("tcp", addr)
		return err
	}, nil, time.Millisecond*30, time.Millisecond*100)

	//tcpConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*TCPConn)

	go func() {
		time.Sleep(time.Millisecond * 100)
		conn.Close()
	}()

	time.Sleep(time.Millisecond * 200)
	request, err := system.Root.Request(agentPid, &requestMsg{})
	assert.NoError(t, err)
	assert.EqualError(t, constant.ErrConnectionClosed, request.Message.(error).Error())
}

func TestGetNextMessageInParts(t *testing.T) {
	addr := "0.0.0.0:6234"
	acceptorProducer := NewTCPAcceptor(addr, func(conn Connector) actor.Producer {
		return func() actor.Actor {
			return &tAgent{
				conn: conn,
				//err:     table.err,
				//message: table.data,
			}
		}
	})
	tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())

	var conn net.Conn
	var err error
	helper.ShouldEventuallyReturn(t, func() error {
		conn, err = net.Dial("tcp", addr)
		return err
	}, nil, time.Millisecond*30, time.Millisecond*100)

	//tcpConn := helper.ShouldEventuallyReceive(t, c, time.Millisecond*100).(*TCPConn)

	part1 := []byte{0x02, 0x00, 0x00, 0x03, 0x01}
	part2 := []byte{0x01, 0x02}

	_, err = conn.Write(part1)
	assert.NoError(t, err)
	go func() {
		time.Sleep(time.Millisecond * 100)
		_, err = conn.Write(part2)
		assert.NoError(t, err)
	}()

	time.Sleep(time.Millisecond * 200)
	request, err := system.Root.Request(agentPid, &requestMsg{})
	assert.NoError(t, err)
	assert.Equal(t, append(part1, part2...), request.Message)
}

func TestTCPAcceptorCloseConn(t *testing.T) {
	addr := "0.0.0.0:7234"
	acceptorProducer := NewTCPAcceptor(addr, func(conn Connector) actor.Producer {
		return func() actor.Actor {
			return &tAgent{
				conn: conn,
				//err:     table.err,
				//message: table.data,
			}
		}
	})
	tcpPID, _ = system.Root.SpawnNamed(actor.PropsFromProducer(acceptorProducer), TCPAcceptorName())

	var conn net.Conn
	var err error
	helper.ShouldEventuallyReturn(t, func() error {
		conn, err = net.Dial("tcp", addr)
		return err
	}, nil, time.Millisecond*30, time.Millisecond*100)

	_ = conn
	system.Root.Poison(tcpPID)
	//conn.Close()

	time.Sleep(time.Second * 1)
	request, err := system.Root.Request(agentPid, &requestMsg{})
	assert.Error(t, err)
	_ = request

	time.Sleep(time.Second * 1)
	//assert.Equal(t, msg1, request.Message)
	//request, err = system.Root.Request(agentPid, &requestMsg{})
	//assert.NoError(t, err)
	//assert.Equal(t, msg2, request.Message)
}
