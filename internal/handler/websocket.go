package handler

import (
	"log"
	"net/http"
	"ws-whatever/internal/domain"
	"ws-whatever/internal/hub"

	"github.com/gorilla/websocket"
)

type WebsocketHandler struct {
	hub *hub.Hub
}

var upgrader = websocket.Upgrader{}

func NewWebsocketHandler(hub *hub.Hub) *WebsocketHandler {
	return &WebsocketHandler{hub: hub}
}

func (h *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Something goes wrong: %v", err)
		return
	}

	client := domain.NewClient(conn, make(chan *domain.Message, 256))
	h.hub.AddClient(client)

	go client.Read(h.hub.Broadcast)
	go client.Write()
}
