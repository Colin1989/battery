package facade

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
)

type Acceptors struct {
	Addr         string
	Certs        [2]string
	AcceptorType constant.AcceptorType
}

type Service interface {
	actor.Actor
	Name() string
	//OnStart(as ActorService)
}

//type ActorService interface {
//}

type ActorHandler func(actor.Context, any)
