package acceptor

import (
	"crypto/tls"
	"fmt"
	"github.com/colin1989/battery/actor"
	"net"
	"sync/atomic"
)

type TCPAcceptor struct {
	addr          string
	running       int32
	listener      net.Listener
	certs         []tls.Certificate
	sessions      map[*actor.PID]net.Conn
	agentProducer connProducer
}

func NewTCPAcceptor(addr string, agentProducer connProducer, certs ...string) actor.Producer {
	return func() actor.Actor {
		certificates := loadCertificate(certs...)
		tcp := newTLSAcceptor(addr, certificates...)
		tcp.agentProducer = agentProducer
		return tcp
	}
}

func newTLSAcceptor(addr string, certs ...tls.Certificate) *TCPAcceptor {
	return &TCPAcceptor{
		addr:     addr,
		certs:    certs,
		sessions: map[*actor.PID]net.Conn{},
	}
}

func TCPAcceptorName() string {
	return "tcp_acceptor"
}

func (ta *TCPAcceptor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		ta.ListenAndServe()
		atomic.StoreInt32(&ta.running, acceptorRunning)
		//blog.Info("socket connector listening at Address %s", zap.String("addr", tcp.GetAddr()))
		go ta.serve(ctx)
	case *actor.Stopped:
		atomic.StoreInt32(&ta.running, acceptorStopped)
		ta.listener.Close()
	default:
		fmt.Printf("unsupported type %T msg : %+v \n", msg, msg)
	}
}

func (ta *TCPAcceptor) ListenAndServe() {
	listener, err := getListener(ta.addr, ta.certs)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %s", err))
	}
	ta.listener = listener
}

func (ta *TCPAcceptor) serve(ctx actor.Context) {
	defer func() {
		atomic.CompareAndSwapInt32(&ta.running, acceptorRunning, acceptorStopped)
	}()
	for atomic.LoadInt32(&ta.running) == acceptorRunning {
		conn, err := ta.listener.Accept()
		if err != nil {
			//blog.Error("Failed to accept TCP connection", zap.Error(err))
			continue
		}

		tcpConn := &TCPConn{
			Conn:       conn,
			remoteAddr: conn.RemoteAddr(),
		}
		ctx.Spawn(actor.PropsFromProducer(ta.agentProducer(tcpConn)))
	}
}
