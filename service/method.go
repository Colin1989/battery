package service

import (
	"github.com/colin1989/battery/net/message"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/colin1989/battery/actor"
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfContext = reflect.TypeOf(new(actor.Context)).Elem()
	//typeOfProtoMsg = reflect.TypeOf(new(proto.Message)).Elem()
)

func isExported(name string) bool {
	w, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(w)
}

// isHandlerMethod decide a method is suitable handler method
func isHandlerMethod(method reflect.Method) bool {
	mt := method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}

	// Method needs two ins: receiver, actor.Context.
	if mt.NumIn() != 2 && mt.NumIn() != 3 {
		return false
	}

	if t1 := mt.In(1); !t1.Implements(typeOfContext) {
		return false
	}

	// Method needs either no out or two outs: interface{}(or []byte), error
	if mt.NumOut() != 0 && mt.NumOut() != 2 {
		return false
	}

	// if mt.NumOut() == 2 && (mt.Out(1) != typeOfError || mt.Out(0) != typeOfBytes && mt.Out(0).Kind() != reflect.Ptr) {
	// 	return false
	// }

	return true
}

func suitableHandlerMethods(typ reflect.Type, nameFunc func(string) string) map[string]*Handler {
	methods := make(map[string]*Handler)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mt := method.Type
		mn := method.Name
		if mn == "Receive" {
			continue
		}
		if isHandlerMethod(method) {
			//raw := false
			// rewrite handler name
			if nameFunc != nil {
				mn = nameFunc(mn)
			}
			var msgType message.Type
			if mt.NumOut() == 0 {
				msgType = message.Notify
			} else {
				msgType = message.Request
			}
			handler := &Handler{
				Method: method,
				//IsRawArg:    raw,
				MessageType: msgType,
			}
			if mt.NumIn() == 3 {
				handler.Type = mt.In(2)
			}
			methods[mn] = handler
		}
	}
	return methods
}
