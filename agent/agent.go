package agent

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/logger"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
	"github.com/colin1989/battery/proto"
	"log/slog"
	"net"
	"reflect"
	"strings"
	"time"
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

	messageEncoder facade.Encoder
	decoder        facade.PacketDecoder
	encoder        facade.PacketEncoder
	serializer     facade.Serializer

	encodedData   []byte // session data encoded as a byte array
	state         int32  // current agent state
	session       *proto.Session
	handshakeData *packet.HandshakeData // handshake data received by the client
}

func NewAgent(conn facade.Connector,
	messageEncoder facade.Encoder,
	decoder facade.PacketDecoder,
	encoder facade.PacketEncoder,
	serializer facade.Serializer) actor.Actor {

	heartbeatTime := time.Minute
	isCompressionEnabled := true
	serializerName := "json"
	once.Do(func() {
		hbdEncode(heartbeatTime, encoder, isCompressionEnabled, serializerName)
		herdEncode(heartbeatTime, encoder, isCompressionEnabled, serializerName)
	})

	return &Agent{
		conn:           conn,
		chSend:         make(chan pendingWrite),
		chDie:          make(chan struct{}),
		messageEncoder: messageEncoder,
		decoder:        decoder,
		encoder:        encoder,
		session:        &proto.Session{Data: make(map[string]string)},
		serializer:     serializer,
	}
}

func (a *Agent) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		logger.Debug("actor started", slog.String("pid", ctx.Self().String()))
		a.ctx = ctx
		a.pid = ctx.Self()
		a.run()
	case *actor.Restarting:
		logger.Debug("actor restarting", slog.String("pid", ctx.Self().String()))
	case *actor.Stopping:
		logger.Debug("actor stopping", slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		logger.Debug("actor stopped", slog.String("pid", ctx.Self().String()))
		a.Close()
	case message.PendingMessage:
		sendPacket(a, msg)
	//case *actor.ReceiveTimeout:
	//	ctx.Stop(ctx.Self())
	case *packet.Packet:
		logger.Debug("actor receive packet", slog.String("pid", ctx.Self().String()),
			slog.String("msg", msg.String()))
		err := processPacket(a, msg)
		if err != nil {
			ctx.Poison(ctx.Self())
			return
		}
	default:
		// TODO Router
		logger.Warn("actor unsupported type",
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
				logger.Error("Failed to write in conn", logger.ErrAttr(err))
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
			logger.CallerStack(r.(error), 1)
		}
		a.ctx.Poison(a.ctx.Self())
	}()

	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			logger.Error("conn receive error",
				slog.String("pid", a.PID()),
				logger.ErrAttr(err))
			a.ctx.Send(a.ctx.Self(), actor.WrapEnvelop(err))
			return
		}
		packets, err := a.decoder.Decode(msg)
		if err != nil {
			logger.Error("Failed to decode message", slog.String("error", err.Error()))
			return
		}

		if len(packets) < 1 {
			logger.Warn("Read no packets", slog.String("pid", a.PID()))
			continue
		}

		// process all packet
		for _, p := range packets {
			a.ctx.Send(a.ctx.Self(), actor.WrapEnvelop(p))
		}
	}
}
