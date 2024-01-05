package service

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/colin1989/battery/logger"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/errors"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/util"
)

type (
	//Handler represents a message.Message's handler's meta information.
	Handler struct {
		Method reflect.Method // method stub
		Type   reflect.Type   // low-level type of method
		// IsRawArg    bool           // whether the data need to serialize
		MessageType message.Type // handler allowed message type (either request or notify)
	}

	ActorService struct {
		facade.App
		Name     string              // name of service
		Type     reflect.Type        // type of the receiver
		Receiver reflect.Value       // receiver of methods for the service
		handlers map[string]*Handler // registered methods

		service facade.Service
	}
)

func NewActorService(service facade.Service, app facade.App) (*ActorService, error) {
	as := &ActorService{}
	as.App = app
	as.service = service
	as.Type = reflect.TypeOf(service)
	as.Receiver = reflect.ValueOf(service)
	if len(service.Name()) != 0 {
		as.Name = service.Name()
	} else {
		as.Name = reflect.Indirect(as.Receiver).Type().Name()
	}

	if err := as.ExtractHandler(); err != nil {
		return nil, err
	}

	return as, nil
}

func (as *ActorService) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		ctx.ActorSystem().Logger().Debug("actor service started", slog.String("name", as.Name),
			slog.String("pid", ctx.Self().String()))
		//as.service.OnStart(as)
	case *actor.Stopping:
		ctx.ActorSystem().Logger().Debug("actor service stopping", slog.String("name", as.Name),
			slog.String("pid", ctx.Self().String()))
	case *actor.Stopped:
		ctx.ActorSystem().Logger().Debug("actor service stopped", slog.String("name", as.Name),
			slog.String("pid", ctx.Self().String()))
	case *message.Message:
		//r.ProcessMessage(ctx, msg)
		//as.service.ProcessMessage(ctx, msg)
		//as.handlers[msg.Route]
		as.handlerMessage(ctx, msg)
	//case *actor.DeadLetterResponse:
	//r.users.Remove(msg.Target)
	//logger.Debug("room DeadLetterResponse", slog.String("pid", ctx.Self().String()))
	default:
		as.service.Receive(ctx)
		ctx.ActorSystem().Logger().Warn("actor service unsupported type",
			slog.String("type", reflect.TypeOf(msg).String()),
			slog.Any("msg", msg),
			slog.String("name", as.Name),
			slog.String("pid", ctx.Self().String()))
	}
}

func (as *ActorService) RegisterHandler(m interface{}, f facade.ActorHandler) {
	//handler := &Handler{
	//	Method: f,
	//	Type:   reflect.TypeOf(m),
	//}
	//if handler.Type.Kind() != reflect.Ptr {
	//	logger.Fatal("need pointer")
	//}
	//method := reflect.TypeOf(m).Elem().Name()
	//as.handlers[strings.ToLower(method)] = handler
}

func (as *ActorService) handlerMessage(ctx actor.Context, msg *message.Message) error {
	route := msg.Route
	handler, ok := as.handlers[route.Method]
	if !ok {
		return fmt.Errorf("pitaya/handler: %s not found", route.String())
	}

	//arg := reflect.New(handler.Type.Elem()).Interface()
	//err := as.Serializer().Unmarshal(msg.Data, arg)
	//if err != nil {
	//	return nil
	//}
	//ctx.Envelope().Message = arg
	//handler.Method(ctx, arg)
	//ctx.Envelope().Message = msg
	msgType, err := getMsgType(msg.Type)
	if err != nil {
		return errors.NewError(err, errors.ErrInternalCode)
	}
	exit, err := handler.ValidateMessageType(msgType)
	if err != nil && exit {
		return errors.NewError(err, errors.ErrBadRequestCode)
	} else if err != nil {
		ctx.ActorSystem().Logger().Warn("invalid message type", logger.ErrAttr(err))
	}

	// First unmarshal the handler arg that will be passed to
	// both handler and pipeline functions
	arg, err := unmarshalHandlerArg(handler, as.Serializer(), msg.Data)
	if err != nil {
		return errors.NewError(err, errors.ErrBadRequestCode)
	}

	args := []reflect.Value{as.Receiver, reflect.ValueOf(ctx)}
	if arg != nil {
		args = append(args, reflect.ValueOf(arg))
	}

	resp, err := Pcall(handler.Method, args)

	ret, err := serializeReturn(as.Serializer(), resp)
	if err != nil {
		return err
	}

	if msgType != message.Notify {
		if err != nil {
			payload, err := util.GetErrorPayload(as.Serializer(), err)
			if err != nil {
				return err
			}
			ctx.Send(ctx.Sender(), actor.WrapResponseEnvelop(msg.ID, payload))
		} else {
			ctx.Send(ctx.Sender(), actor.WrapResponseEnvelop(msg.ID, ret))
		}
	}

	return nil
}

// ExtractHandler extract the set of methods from the
// receiver value which satisfy the following conditions:
// - exported method of exported type
// - one or two arguments
// - the first argument is context.Context
// - the second argument (if it exists) is []byte or a pointer
// - zero or two outputs
// - the first output is [] or a pointer
// - the second output is an error
func (as *ActorService) ExtractHandler() error {
	typeName := reflect.Indirect(as.Receiver).Type().Name()
	if typeName == "" {
		return errors.Errors("no service name for type " + as.Type.String())
	}
	if !isExported(typeName) {
		return errors.Errors("type " + typeName + " is not exported")
	}

	nameFunc := strings.ToLower
	// Install the methods
	as.handlers = suitableHandlerMethods(as.Type, nameFunc)

	if len(as.handlers) == 0 {
		str := ""
		// To help the user, see if a pointer receiver would work.
		method := suitableHandlerMethods(reflect.PtrTo(as.Type), nameFunc)
		if len(method) != 0 {
			str = "type " + as.Name + " has no exported methods of handler type (hint: pass a pointer to value of that type)"
		} else {
			str = "type " + as.Name + " has no exported methods of handler type"
		}
		return errors.Errors(str)
	}

	return nil
}
