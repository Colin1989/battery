package battery

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/agent"
	"github.com/colin1989/battery/constant"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
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
	system *actor.ActorSystem
	actors actor.PIDSet // actor was spawn by root context
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

func (app *Application) Start() {
	defer func() {
		if r := recover(); r != nil {
			//blog.Error("Start recover", zap.Any("reason", r))
			//blog.Error("Start recover", zap.Stack("recover"))
		}
	}()

	//defer func() {
	//	blog.Flush()
	//}()

	amPID, err := app.system.Root.SpawnNamed(actor.PropsFromProducer(func() actor.Actor {
		return agent.NewAgentManager()
	}), constant.AgentManager)
	if err != nil {
		panic(err)
	}
	app.actors.Add(amPID)

	atomic.AddInt32(&app.running, 1)

	// print version info
	fmt.Print(GetLOGO())

	//app.listen()

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
	fmt.Println(fmt.Sprintf("actor system is stopping ..."))
	app.actors.ForEach(func(_ int, pid *actor.PID) {
		app.system.Root.Poison(pid)
	})
	app.system.Shutdown()
	fmt.Println(fmt.Sprintf("actor system is stopped"))
}
