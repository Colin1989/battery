package blog

import (
	"log/slog"
	"reflect"
)

func ErrAttr(err error) slog.Attr {
	return slog.String("err", err.Error())
}

func TypeAttr(v interface{}) slog.Attr {
	return slog.String("type", reflect.TypeOf(v).String())
}
