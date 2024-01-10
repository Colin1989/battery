package battery

import (
	"log/slog"

	"github.com/colin1989/battery/errors"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/blog"
	"github.com/colin1989/battery/net/codec"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/serializer/json"
)

func NewApp(opts ...Option) *Application {
	system := actor.NewActorSystem(actor.WithLoggerFactory(func(system *actor.ActorSystem) *slog.Logger {
		return blog.Logger()
	}))
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

// Error creates a new error with a code, message and metadata
func Error(err error, code string, metadata ...map[string]string) *errors.Error {
	return errors.NewError(err, code, metadata...)
}
