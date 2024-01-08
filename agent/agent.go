package agent

import (
	"log/slog"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/blog"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
	"github.com/colin1989/battery/proto"
)

type pendingWrite struct {
	data []byte
	err  error
}

type Agent struct {
	ctx    actor.Context
	pid    *actor.PID
	conn   facade.Connector
	chSend chan pendingWrite // push message queue
	chDie  chan struct{}     // wait for close

	app facade.App

	encodedData   []byte // session data encoded as a byte array
	state         int32  // current agent state
	session       *proto.Session
	handshakeData *packet.HandshakeData // handshake data received by the client
}

func NewAgent(conn facade.Connector, app facade.App) actor.Actor {

	heartbeatTime := time.Minute
	isCompressionEnabled := true
	serializerName := "json"
	once.Do(func() {
		hbdEncode(heartbeatTime, app.Encoder(), isCompressionEnabled, serializerName)
		herdEncode(heartbeatTime, app.Encoder(), isCompressionEnabled, serializerName)
	})

	return &Agent{
		conn:    conn,
		chSend:  make(chan pendingWrite),
		chDie:   make(chan struct{}),
		app:     app,
		session: &proto.Session{Data: make(map[string]string)},
	}
}

func (a *Agent) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		blog.Debug("actor started", slog.String("pid", ctx.Self().String()))
		a.ctx = ctx
		a.pid = ctx.Self()
		a.run()
	case *actor.Restarting:
		blog.Debug("actor restarting", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		blog.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		blog.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		a.Close()
	case message.PendingMessage:
		sendPacket(a, msg)
	//case *actor.ReceiveTimeout:
	//	ctx.Stop(ctx.Self())
	case *packet.Packet:
		blog.Debug("actor receive packet", slog.String("pid", ctx.Self().String()),
			slog.String("msg", msg.String()))
		err := processPacket(a, msg)
		if err != nil {
			ctx.Poison(ctx.Self())
			return
		}
	default:
		// TODO Router
		blog.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}

func (a *Agent) PID() string {
	return a.pid.String()
}

func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

// IPVersion returns the remote address ip version.
// net.TCPAddr and net.UDPAddr implementations of String()
// always construct result as <ip>:<port> on both
// ipv4 and ipv6. Also, to see if the ip is ipv6 they both
// check if there is a colon on the string.
// So checking if there are more than one colon here is safe.
func (a *Agent) IPVersion() string {
	version := constant.IPv4

	ipPort := a.RemoteAddr().String()
	if strings.Count(ipPort, ":") > 1 {
		version = constant.IPv6
	}

	return version
}

func (a *Agent) SetLastAt() {

}

func (a *Agent) Close() {
	close(a.chDie)
	a.conn.Close()
}

func (a *Agent) send(data []byte) error {
	a.chSend <- pendingWrite{
		data: data,
		err:  nil,
	}
	return nil
}

func (a *Agent) run() {
	go a.write()
	go a.read()
}

func (a *Agent) write() {
	defer func() {
		close(a.chSend)
	}()

	for {
		select {
		case pWrite, ok := <-a.chSend:
			if !ok {
				return
			}
			if _, err := a.conn.Write(pWrite.data); err != nil {
				blog.Error("Failed to write in conn", blog.ErrAttr(err))
				return
			}
		case <-a.chDie:
			return
		}
	}
}

func (a *Agent) read() {
	defer func() {
		if r := recover(); r != nil {
			blog.CallerStack(r.(error), 1)
		}
		a.ctx.Poison(a.ctx.Self())
	}()

	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			blog.Error("conn receive error",
				slog.String("pid", a.PID()),
				blog.ErrAttr(err))
			a.ctx.Send(a.ctx.Self(), actor.WrapEnvelope(err))
			return
		}
		packets, err := a.app.Decoder().Decode(msg)
		if err != nil {
			blog.Error("Failed to decode message", slog.String("error", err.Error()))
			return
		}

		if len(packets) < 1 {
			blog.Warn("Read no packets", slog.String("pid", a.PID()))
			continue
		}

		// process all packet
		for _, p := range packets {
			a.ctx.Send(a.ctx.Self(), actor.WrapEnvelope(p))
		}
	}
}
