package domain

import "log"

type Hub struct {
	Clients   ClientList
	Join      chan *Client
	Leave     chan *Client
	Broadcast chan *Message
}

func NewHub() *Hub {
	return &Hub{
		Clients:   make(ClientList),
		Join:      make(chan *Client),
		Leave:     make(chan *Client),
		Broadcast: make(chan *Message),
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

func (h *Hub) AddClient(c *Client) {
	// dunno about this, like wtf?
	h.Join <- c
}
