package domain

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	ConnID string
	Conn   *ws.Conn
	Send   chan *Message
	Hub    *Hub
}

func NewClient(conn *ws.Conn, send chan *Message, hub *Hub) *Client {
	return &Client{
		ConnID: uuid.New().String(),
		Conn:   conn,
		Send:   send,
		Hub:    hub,
	}
}

func (c *Client) Read() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("ERROR: %v", err)
			}
			break
		}

		message := Message{}

		if err := json.Unmarshal(p, &message); err != nil {
			log.Printf("Unmarshal message goes wrong: %v", err)
			break
		}

		if message.ID == nil {
			id := uuid.New().String()
			message.ID = &id
		}

		log.Printf("Message: %v", message)
	}
}
func (c *Client) Write() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				break
			}
			if err := c.Conn.WriteJSON(msg); err != nil {
				log.Printf("%v", err)
				break
			}
		}
	}
}
