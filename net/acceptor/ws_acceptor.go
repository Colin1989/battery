package acceptor

import (
	"crypto/tls"
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
)

type WSAcceptor struct {
	addr          string
	listener      net.Listener
	upgrade       *websocket.Upgrader
	certs         []tls.Certificate
	ctx           actor.Context
	agentProducer connProducer
}

func (wa *WSAcceptor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		wa.ListenAndServe()
		wa.ctx = ctx
		//blog.Info("socket connector listening at Address %s", zap.String("addr", tcp.GetAddr()))
	case *actor.Stopped:
		wa.listener.Close()
	default:
		fmt.Printf("unsupported type %T msg : %+v \n", msg, msg)
	}
}

func NewWsAcceptor(addr string, agentProducer connProducer, certs ...string) actor.Producer {
	return func() actor.Actor {
		w := &WSAcceptor{
			addr:          addr,
			certs:         loadCertificate(certs...),
			agentProducer: agentProducer,
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
}

func WSAcceptorName() string {
	return "ws_acceptor"
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
	wa.ctx.Spawn(actor.PropsFromProducer(wa.agentProducer(NewWSConn(conn))))
}
