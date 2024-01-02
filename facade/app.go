package facade

type App interface {
	MessageEncoder() MessageEncoder
	Decoder() PacketDecoder
	Encoder() PacketEncoder
	Serializer() Serializer
}
