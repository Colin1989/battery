package errors

import (
	e "errors"
	"fmt"
)

func Errors(text string) error {
	return e.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

var (
	ErrProduceApplication             = Errors("produce app is failed")
	ErrProfileFilePathNil             = Errors("profile file path is nil.")
	ErrProfileNodeIdNil               = Errors("NodeId is nil.")
	ErrConnectorProducerFuncNil       = Errors("connector producer func is nil")
	ErrInvalidCertificates            = Errors("invalid certificates")
	ErrIncorrectNumberOfCertificates  = Errors("certificates must be exactly two")
	ErrReceivedMsgSmallerThanExpected = Errors("received less data than expected, EOF?")
	ErrReceivedMsgBiggerThanExpected  = Errors("received more data than expected")
	ErrConnectionClosed               = Errors("client connection closed")
	ErrInvalidMsg                     = Errors("invalid message type provided")
	ErrNotifyOnRequest                = Errors("tried to notify a request route")
	ErrRequestOnNotify                = Errors("tried to request a notify route")
	ErrReplyShouldBeNotNull           = Errors("reply must not be null")
	ErrWrongValueType                 = Errors("protobuf: convert on wrong type value")
	ErrRouteFieldCantEmpty            = Errors("route field can not be empty")
	ErrInvalidRoute                   = Errors("invalid route")

	ErrRPCRequestTimeout       = Errors("rpc client: request timeout")
	ErrRPCClientNotInitialized = Errors("RPC client is not running")
)
