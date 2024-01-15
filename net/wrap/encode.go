package wrap

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
	"github.com/colin1989/battery/util"
)

func WrapPushEnvelop(route string, v interface{}) *actor.MessageEnvelope {
	route1, _ := message.DecodeRoute(route)
	m := message.PendingMessage{
		Typ:     message.Push,
		Route:   route1,
		Payload: v,
	}
	return &actor.MessageEnvelope{
		Header:  nil,
		Message: m,
	}
}

func WrapResponseEnvelop(mid uint, v interface{}) *actor.MessageEnvelope {
	m := message.PendingMessage{
		Typ:     message.Response,
		Mid:     mid,
		Payload: v,
	}
	return &actor.MessageEnvelope{
		Header:  nil,
		Message: m,
	}
}

func WrapBroadcast(app facade.App, route string, v interface{}) *actor.MessageEnvelope {
	route1, _ := message.DecodeRoute(route)
	pendingMessage := message.PendingMessage{
		Typ:     message.Push,
		Route:   route1,
		Payload: v,
	}

	payload, _ := util.SerializeOrRaw(app.Serializer(), pendingMessage.Payload)
	// construct message and encode
	m := &message.Message{
		Type:  pendingMessage.Typ,
		Data:  payload,
		Route: pendingMessage.Route,
		ID:    pendingMessage.Mid,
		Err:   pendingMessage.Err,
	}

	em, err := app.MessageEncoder().Encode(m)
	if err != nil {
		//blog.Error("actor send client", slog.String("pid", a.PID()),
		//	blog.ErrAttr(err))
		return nil
	}
	p, err := app.Encoder().Encode(packet.Data, em)
	if err != nil {
		//blog.Error("actor send client", slog.String("pid", a.PID()),
		//	blog.ErrAttr(err))
		return nil
	}

	return &actor.MessageEnvelope{
		Header:  nil,
		Message: &message.BroadcastMessage{P: p},
	}
}
