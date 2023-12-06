package acceptor

import (
	"crypto/tls"
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/logger"
	"github.com/colin1989/battery/message"
	"github.com/gorilla/websocket"
	"log/slog"
	"net"
	"net/http"
	"reflect"
)

type WSAcceptor struct {
	addr     string
	listener net.Listener
	upgrade  *websocket.Upgrader
	certs    []tls.Certificate
	ctx      actor.Context
}

func NewWSAcceptor(addr string, certs ...string) *WSAcceptor {
	w := &WSAcceptor{
		addr:  addr,
		certs: loadCertificate(certs...),
		upgrade: &websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	return w
}

func (wa *WSAcceptor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		wa.ctx = ctx
		go wa.ListenAndServe()
		//blog.Info("socket connector listening at Address %s", zap.String("addr", tcp.GetAddr()))
	case *actor.Stopping:
	case *actor.Stopped:
		wa.ctx = nil
		wa.listener.Close()
	default:
		logger.Warn("actor unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg))
	}
}

func (wa *WSAcceptor) ListenAndServe() {
	var err error
	wa.listener, err = getListener(wa.addr, wa.certs)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %s", err))
	}

	//blog.Info("Websocket connector listening at Address %s", zap.String("addr", wa.GetAddr()))

	http.Serve(wa.listener, wa)
}

func (wa *WSAcceptor) Stop() {
	if err := wa.listener.Close(); err != nil {
		//blog.Error("Failed to stop", zap.Error(err))
	}
}

func (wa *WSAcceptor) GetAddr() string {
	if wa.listener != nil {
		return wa.listener.Addr().String()
	}
	return ""
}

func (wa *WSAcceptor) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn, err := wa.upgrade.Upgrade(writer, request, nil)
	if err != nil {
		//blog.Error("Upgrade failure", zap.String("URI", request.RequestURI), zap.Error(err))
		return
	}

	connector := NewWSConn(conn)
	system := wa.ctx.ActorSystem()
	wa.ctx.ActorSystem().Root.Send(system.NewLocalPID(constant.AgentManager), actor.WrapEnvelop(&message.NewAgent{Conn: connector}))
}
