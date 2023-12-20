package battery

import (
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/net/codec"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/serializer/json"
)

func NewApp(opts ...Option) *Application {
	system := actor.NewActorSystem()
	app := &Application{
		debug:      false,
		dieChan:    make(chan bool),
		serverMode: Cluster,
		system:     system,

		messageEncoder: message.NewMessagesEncoder(true),
		decoder:        codec.NewPomeloPacketDecoder(),
		encoder:        codec.NewPomeloPacketEncoder(),
		serializer:     json.NewSerializer(),
	}

	//app.AppConfig = profile.NewDefaultAppConfig()
	//logger.SetNodeLogger("node")

	for _, opt := range opts {
		if err := opt(app); err != nil {
			panic(err)
		}
	}

	return app
}
