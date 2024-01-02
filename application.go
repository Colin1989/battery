package battery

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/logger"
	"github.com/colin1989/battery/service"
)

// ServerMode represents a server mode
type ServerMode byte

const (
	_ ServerMode = iota
	// Cluster represents a server running with connection to other servers
	Cluster
	// Standalone represents a server running without connection to other servers
	Standalone
)

type Application struct {
	//profile.AppConfig
	debug      bool
	running    int32
	isFrontend bool
	dieChan    chan bool
	serverMode ServerMode
	//startTime  btime.Time

	messageEncoder facade.MessageEncoder
	decoder        facade.PacketDecoder
	encoder        facade.PacketEncoder
	serializer     facade.Serializer

	system   *actor.ActorSystem
	services []facade.Service
	actors   actor.PIDSet // actor was spawn by root context
}

func (app *Application) Register(s facade.Service) {
	app.services = append(app.services, s)
}

func (app *Application) IsFrontend() bool {
	return app.isFrontend
}

func (app *Application) IsRunning() bool {
	return atomic.LoadInt32(&app.running) > 0
}

func (app *Application) DieChan() chan bool {
	return app.dieChan
}

func (app *Application) NodeMode() ServerMode {
	return app.serverMode
}

func (app *Application) Shutdown() {
	select {
	case <-app.dieChan: // 避免重复关闭
	default:
		close(app.dieChan)
	}
}

func (app *Application) addService(s facade.Service) {
	as, err := service.NewActorService(s, app)
	if err != nil {
		logger.Fatal("addService", logger.ErrAttr(err))
	}
	props := actor.PropsFromProducer(func() actor.Actor {
		return as
	})
	pid, err := app.system.Root.SpawnNamed(props, s.Name())
	if err != nil {
		logger.Fatal("new service", slog.Any("service", s.Name()), logger.ErrAttr(err))
	}
	app.actors.Add(pid)
}

func (app *Application) Start() {
	defer func() {
		if r := recover(); r != nil {
			logger.CallerStack(r.(error), 1)
		}
	}()

	//defer func() {
	//	blog.Flush()
	//}()

	atomic.AddInt32(&app.running, 1)

	// print version info
	fmt.Print(GetLOGO())

	for _, service := range app.services {
		app.addService(service)
	}

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	select {
	case <-app.dieChan:
		//blog.Warn("the app will shutdown in app few seconds")
	case s := <-sg:
		_ = s
		//blog.Warn("shutting down...", zap.Stringer("signal", s))
		close(app.dieChan)
	}

	//blog.Warn("server is stopping...")
	app.shutdownActorSystem()

	// stop status
	atomic.StoreInt32(&app.running, 0)
}

func (app *Application) shutdownActorSystem() {
	logger.Info("actor system is stopping ...")
	app.actors.ForEach(func(_ int, pid *actor.PID) {
		app.system.Root.Poison(pid)
	})
	app.system.Shutdown()
	logger.Info("actor system is stopped")
}
