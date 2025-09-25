package main

import (
	"html/template"
	"log"
	"net/http"

	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{}

func main() {
	mux := http.NewServeMux()
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Something goes wrong: %v", err)
			return
		}

		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
					log.Printf("ERROR: %v", err)
				}
				break
			}

			log.Println(string(payload))
		}
	})

	log.Println("Server running on port :6969")
	if err := http.ListenAndServe(":6969", mux); err != nil {
		log.Fatal(err)
	}
}
