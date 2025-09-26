package main

import (
	"html/template"
	"log"
	"net/http"
	"ws-whatever/internal/handler"
	"ws-whatever/internal/hub"
)

func main() {
	mux := http.NewServeMux()
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))

	hub := hub.NewHub()

	go hub.Run()

	httpHandler := handler.NewHTTPHandler(tmpl)
	websocketHandler := handler.NewWebsocketHandler(hub)

	mux.Handle("GET /", httpHandler)
	mux.Handle("GET /ws", websocketHandler)

	log.Println("Server running on port :6969")
	if err := http.ListenAndServe(":6969", mux); err != nil {
		log.Fatal(err)
	}
}
