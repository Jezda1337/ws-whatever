package ws

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	ID      string
	Conn    *ws.Conn
	Manager *Manager
	Send    chan []byte
	Room    string
}

func NewClient(conn *ws.Conn, m *Manager) *Client {
	id := uuid.New().String()

	return &Client{ID: id, Conn: conn, Manager: m, Send: make(chan []byte)}
}

func (c *Client) ReadMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("Error reading messages: %v", err)
			}
			break
		}

		event := Event{}
		if err := json.Unmarshal(p, &event); err != nil {
			log.Printf("Invalid message: %v", err)
			continue
		}

		switch event.Type {
		case "send_message":
			outgoing := Event{
				Type:    "new_message",
				Payload: event.Payload,
				From:    c.ID,
			}
			data, _ := json.Marshal(outgoing)

			c.Manager.RLock()
			for client := range c.Manager.Clients {
				if client.Room == c.Room {
					client.Send <- data
				}
			}
			c.Manager.RUnlock()

		case "change_room":
			// TODO: move client to another room (next step)
			c.Room = event.Payload
		}
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				if err := c.Conn.WriteMessage(ws.CloseMessage, nil); err != nil {
					log.Printf("Connection closed: %v", err)
				}
				return
			}
			if err := c.Conn.WriteMessage(ws.TextMessage, message); err != nil {
				log.Printf("Failed to send a message: %v", err)
			}

			// if err := c.Conn.WriteMe(message); err != nil {
			// 	log.Printf("Failed to send a message: %v", err)
			// 	return
			// }
			log.Println("Message sent")
		}
	}
}
