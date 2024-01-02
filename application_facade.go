package battery

import "github.com/colin1989/battery/facade"

func (app *Application) MessageEncoder() facade.MessageEncoder {
	return app.messageEncoder
}

func (app *Application) Decoder() facade.PacketDecoder {
	return app.decoder
}

func (app *Application) Encoder() facade.PacketEncoder {
	return app.encoder
}

func (app *Application) Serializer() facade.Serializer {
	return app.serializer
}
