package constant

import (
	"errors"
	"fmt"
)

func Error(text string) error {
	return errors.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf(format, a...))
}

var (
	ErrProduceApplication             = Error("produce app is failed")
	ErrProfileFilePathNil             = Error("profile file path is nil.")
	ErrProfileNodeIdNil               = Error("NodeId is nil.")
	ErrConnectorProducerFuncNil       = Error("connector producer func is nil")
	ErrInvalidCertificates            = Error("invalid certificates")
	ErrIncorrectNumberOfCertificates  = Error("certificates must be exactly two")
	ErrReceivedMsgSmallerThanExpected = Error("received less data than expected, EOF?")
	ErrReceivedMsgBiggerThanExpected  = Error("received more data than expected")
	ErrConnectionClosed               = Error("client connection closed")
)
