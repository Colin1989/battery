package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/colin1989/battery"
	"github.com/colin1989/battery/blog"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
)

func main() {
	blog.NewLogger(blog.LogConfig{
		LogLevel: "error",
		LogPath:  "log/log.log",
		//MaxSize:    0,
		//MaxAge:     0,
		//MaxBackups: 0,
	})
	app := battery.NewApp(battery.WithGate([]facade.Acceptors{
		{"0.0.0.0:2250", [2]string{}, constant.AcceptorTypeWS},
	}))
	app.Register(NewRoomService(app))

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	go http.ListenAndServe(":2251", nil)
	blog.Infof("http run. http://%s", "localhost:2251/web/")

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	app.Start()
	defer app.Shutdown()
}
