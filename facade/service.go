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
	//actor.Actor
	Name() string
	App() App
	OnStart(ctx actor.Context)
	OnDestroy(ctx actor.Context)
}

//type ActorService interface {
//	Context() *actor.Context
//}

type ActorHandler func(actor.Context, any)
