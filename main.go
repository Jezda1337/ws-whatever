package main

import (
	"html/template"
	"log"
	"net/http"
	"ws-whatever/ws"

	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	mux := http.NewServeMux()
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	// logger := utils.NewLogger()

	m := ws.NewManager()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Something goes wrong with upgrading protocol: %v", err)
			return
		}

		c := ws.NewClient(conn, m)
		c.Manager.AddClient(c)

		go c.ReadMessages()
		go c.WriteMessages()
	})

	// serving static files
	fs := http.FileServer(http.Dir("web/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	log.Println("Server running on port :6969")
	log.Fatal(http.ListenAndServe(":6969", mux))
}
