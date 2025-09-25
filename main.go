package main

import (
	"html/template"
	"log"
	"net/http"
	"ws-whatever/domain"

	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{}

func main() {
	mux := http.NewServeMux()
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	hub := domain.NewHub()

	go hub.Run()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Something goes wrong: %v", err)
			return
		}

		client := domain.NewClient(conn, make(chan *domain.Message, 256), hub)
		hub.AddClient(client)

		go client.Read()
		go client.Write()
	})

	log.Println("Server running on port :6969")
	if err := http.ListenAndServe(":6969", mux); err != nil {
		log.Fatal(err)
	}
}
