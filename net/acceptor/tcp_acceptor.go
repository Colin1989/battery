package acceptor

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"sync/atomic"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/blog"
)

type TCPAcceptor struct {
	addr     string
	running  int32
	listener net.Listener
	certs    []tls.Certificate
	ctx      actor.Context
}

func NewTCPAcceptor(addr string, certs ...string) actor.Actor {
	tcp := &TCPAcceptor{
		addr:  addr,
		certs: loadCertificate(certs...),
	}
	return tcp
}

func (ta *TCPAcceptor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		ta.ctx = ctx
		go ta.ListenAndServe()
		atomic.StoreInt32(&ta.running, acceptorRunning)
		//blog.Info("socket connector listening at Address %s", zap.String("addr", tcp.GetAddr()))
	case *actor.Stopping:
	case *actor.Stopped:
		atomic.StoreInt32(&ta.running, acceptorStopped)
		ta.listener.Close()
	default:
		blog.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}

func (ta *TCPAcceptor) ListenAndServe() {
	listener, err := getListener(ta.addr, ta.certs)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %s", err))
	}
	ta.listener = listener

	ta.serve()
}

func (ta *TCPAcceptor) serve() {
	defer func() {
		atomic.CompareAndSwapInt32(&ta.running, acceptorRunning, acceptorStopped)
	}()
	for atomic.LoadInt32(&ta.running) == acceptorRunning {
		conn, err := ta.listener.Accept()
		if err != nil {
			//blog.Error("Failed to accept TCP connection", zap.Error(err))
			continue
		}

		connector := &TCPConn{
			Conn:       conn,
			remoteAddr: conn.RemoteAddr(),
		}
		ta.ctx.ActorSystem().Root.Send(ta.ctx.Parent(), actor.WrapEnvelope(connector))
	}
}
