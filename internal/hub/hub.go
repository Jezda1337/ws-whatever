package hub

import (
	"log"
	"ws-whatever/internal/domain"
)

type Hub struct {
	Clients   domain.ClientList
	Join      chan *domain.Client
	Leave     chan *domain.Client
	Broadcast chan *domain.Message
}

func NewHub() *Hub {
	return &Hub{
		Clients:   make(domain.ClientList),
		Join:      make(chan *domain.Client),
		Leave:     make(chan *domain.Client),
		Broadcast: make(chan *domain.Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Join:
			h.Clients[c] = true
			log.Printf("CLIENT JUST JOIN: %v", c.ConnID)
		case c := <-h.Leave:
			delete(h.Clients, c)
			log.Printf("CLIENT JUST LEAVE: %v", &c.ConnID)
		case msg := <-h.Broadcast:
			for c := range h.Clients {
				select {
				case c.Send <- msg:
				default:
					close(c.Send)
				}
			}
		}
	}
}

func (h *Hub) AddClient(c *domain.Client) {
	h.Join <- c
}
