package util

import (
	"github.com/colin1989/battery/errors"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/proto"
)

// SerializeOrRaw serializes the interface if its not an array of bytes already
func SerializeOrRaw(serializer facade.Serializer, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	data, err := serializer.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetErrorPayload creates and serializes an error payload
func GetErrorPayload(serializer facade.Serializer, err error) ([]byte, error) {
	code := errors.ErrUnknownCode
	msg := err.Error()
	metadata := map[string]string{}
	if val, ok := err.(*errors.Error); ok {
		code = val.Code
		metadata = val.Metadata
	}
	errPayload := &proto.Error{
		Code: code,
		Msg:  msg,
	}
	if len(metadata) > 0 {
		errPayload.Metadata = metadata
	}
	return SerializeOrRaw(serializer, errPayload)
}
