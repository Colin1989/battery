package battery

import (
	"github.com/colin1989/battery/actor"
)

func NewApp(opts ...Option) *Application {
	system := actor.NewActorSystem()
	app := &Application{
		debug:      false,
		dieChan:    make(chan bool),
		serverMode: Cluster,
		system:     system,
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
