package actor

type messageHeader map[string]string

func (header messageHeader) Get(key string) string {
	return header[key]
}

func (header messageHeader) Set(key string, value string) {
	header[key] = value
}

func (header messageHeader) Keys() []string {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	return keys
}

func (header messageHeader) Length() int {
	return len(header)
}

func (header messageHeader) ToMap() map[string]string {
	mp := make(map[string]string)
	for k, v := range header {
		mp[k] = v
	}
	return mp
}

type ReadonlyMessageHeader interface {
	Get(key string) string
	Keys() []string
	Length() int
	ToMap() map[string]string
}

type MessageEnvelope struct {
	Header  messageHeader
	Message interface{}
	Sender  *PID
}

func (envelope *MessageEnvelope) GetHeader(key string) string {
	if envelope.Header == nil {
		return ""
	}
	return envelope.Header.Get(key)
}

func (envelope *MessageEnvelope) SetHeader(key string, value string) {
	if envelope.Header == nil {
		envelope.Header = make(map[string]string)
	}
	envelope.Header.Set(key, value)
}

var EmptyMessageHeader = make(messageHeader)

func UnwrapEnvelope(message interface{}) (ReadonlyMessageHeader, interface{}, *PID) {
	if env, ok := message.(*MessageEnvelope); ok {
		return env.Header, env.Message, env.Sender
	}
	return nil, message, nil
}

func WrapEnvelope(message interface{}) *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: message,
		Sender:  nil,
	}
}

func WrapEnvelopWithSender(message interface{}, sender *PID) *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: message,
		Sender:  sender,
	}
}

func UnwrapEnvelopeMessage(message interface{}) interface{} {
	if env, ok := message.(*MessageEnvelope); ok {
		return env.Message
	}
	return message
}
