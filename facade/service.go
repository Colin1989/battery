package facade

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/net/message"
)

type Acceptors struct {
	Addr         string
	Certs        [2]string
	AcceptorType constant.AcceptorType
}

type Service interface {
	actor.Actor
	Name() string
	ProcessMessage(actor.Context, *message.Message)
}
