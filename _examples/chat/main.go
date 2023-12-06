package main

import (
	"github.com/colin1989/battery"
	"net/http"
)

func main() {
	app := battery.NewApp(battery.WithWSAcceptor("0.0.0.0:3250"))

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	go http.ListenAndServe(":3251", nil)

	app.Start()
	defer app.Shutdown()
}
