package service

import (
	"fmt"
	serialize "github.com/colin1989/battery/serializer"
	"github.com/colin1989/battery/util"
	"reflect"

	"github.com/colin1989/battery/errors"
	"github.com/colin1989/battery/logger"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/proto"
)

func getMsgType(msgTypeIface interface{}) (message.Type, error) {
	var msgType message.Type
	if val, ok := msgTypeIface.(message.Type); ok {
		msgType = val
	} else if val, ok := msgTypeIface.(proto.MsgType); ok {
		msgType = ConvertProtoToMessageType(val)
	} else {
		return msgType, errors.ErrInvalidMsg
	}
	return msgType, nil
}

// ValidateMessageType validates a given message type against the handler's one
// and returns an error if it is a mismatch and a boolean indicating if the caller should
// exit in the presence of this error or not.
func (h *Handler) ValidateMessageType(msgType message.Type) (exitOnError bool, err error) {
	if h.MessageType != msgType {
		switch msgType {
		case message.Request:
			err = errors.ErrRequestOnNotify
			exitOnError = true

		case message.Notify:
			err = errors.ErrNotifyOnRequest
		}
	}
	return
}

// ConvertProtoToMessageType converts a protos.MsgType to a message.Type
func ConvertProtoToMessageType(protoMsgType proto.MsgType) message.Type {
	var msgType message.Type
	switch protoMsgType {
	case proto.MsgType_MsgRequest:
		msgType = message.Request
	case proto.MsgType_MsgNotify:
		msgType = message.Notify
	}
	return msgType
}

func unmarshalHandlerArg(handler *Handler, serializer serialize.Serializer, payload []byte) (interface{}, error) {
	//if handler.IsRawArg {
	//	return payload, nil
	//}

	var arg interface{}
	if handler.Type != nil {
		arg = reflect.New(handler.Type.Elem()).Interface()
		err := serializer.Unmarshal(payload, arg)
		if err != nil {
			return nil, err
		}
	}
	return arg, nil
}

func serializeReturn(ser serialize.Serializer, ret interface{}) ([]byte, error) {
	res, err := util.SerializeOrRaw(ser, ret)
	if err != nil {
		logger.Error("Failed to serialize return", logger.ErrAttr(err))
		res, err = util.GetErrorPayload(ser, err)
		if err != nil {
			logger.Error("cannot serialize message and respond to the client", logger.ErrAttr(err))
			return nil, err
		}
	}
	return res, nil
}

// Pcall calls a method that returns an interface and an error and recovers in case of panic
func Pcall(method reflect.Method, args []reflect.Value) (rets interface{}, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			// Try to use logger from context here to help trace error cause
			// stackTrace := debug.Stack()
			// stackTraceAsRawStringLiteral := strconv.Quote(string(stackTrace))
			// log := getLoggerFromArgs(args)
			// log.Errorf("panic - pitaya/dispatch: methodName=%s panicData=%v stackTrace=%s", method.Name, rec, stackTraceAsRawStringLiteral)

			if s, ok := rec.(string); ok {
				err = errors.Errors(s)
			} else {
				err = fmt.Errorf("rpc call internal error - %s: %v", method.Name, rec)
			}
			logger.CallerStack(err, 1)
		}
	}()

	r := method.Func.Call(args)
	// r can have 0 length in case of notify handlers
	// otherwise it will have 2 outputs: an interface and an error
	if len(r) == 2 {
		if v := r[1].Interface(); v != nil {
			err = v.(error)
		} else if !r[0].IsNil() {
			rets = r[0].Interface()
		} else {
			err = errors.ErrReplyShouldBeNotNull
		}
	}
	return
}
