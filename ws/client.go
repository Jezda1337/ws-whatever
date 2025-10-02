package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type Client struct {
	ID      string
	UserID  int
	Conn    *ws.Conn
	Manager *Manager
	Send    chan []byte
	RoomID  int
}

func NewClient(conn *ws.Conn, m *Manager, userID int) *Client {
	id := uuid.New().String()

	return &Client{
		ID:      id,
		UserID:  userID,
		Conn:    conn,
		Manager: m,
		Send:    make(chan []byte, 256),
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				c.Manager.logger.Error("unexpected websocket close", "error", err, "userID", c.UserID)
			}
			break
		}

		event := Event{}
		if err := json.Unmarshal(p, &event); err != nil {
			c.Manager.logger.Warn("invalid message format", "error", err, "userID", c.UserID)
			continue
		}

		if err := c.handleEvent(event); err != nil {
			c.Manager.logger.Error("event handling error", "type", event.Type, "error", err, "userID", c.UserID)
			c.sendError(err.Error())
		}
	}
}

func (c *Client) WriteMessages() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Manager.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				if err := c.Conn.WriteMessage(ws.CloseMessage, nil); err != nil {
					c.Manager.logger.Error("failed to send close message", "error", err, "userID", c.UserID)
				}
				return
			}
			if err := c.Conn.WriteMessage(ws.TextMessage, message); err != nil {
				c.Manager.logger.Error("failed to send message", "error", err, "userID", c.UserID)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleEvent(event Event) error {
	switch event.Type {
	case "send_message":
		return c.handleSendMessage(event.Payload)
	case "join_room":
		return c.handleJoinRoom(event.Payload)
	case "typing":
		return c.handleTyping()
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (c *Client) handleSendMessage(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var msg SendMessagePayload
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	if msg.Content == "" {
		return errors.New("message content cannot be empty")
	}

	if c.RoomID == 0 {
		return errors.New("must join a room before sending messages")
	}

	message := Message{
		RoomID:    c.RoomID,
		SenderID:  c.UserID,
		Content:   msg.Content,
		ReplyToID: msg.ReplyToID,
	}

	if err := c.Manager.db.Create(&message).Error; err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	outgoing := Event{
		Type: "new_message",
		Payload: NewMessagePayload{
			ID:        message.ID,
			RoomID:    message.RoomID,
			SenderID:  message.SenderID,
			Content:   message.Content,
			ReplyToID: message.ReplyToID,
			CreatedAt: message.CreatedAt,
		},
	}

	data, err = json.Marshal(outgoing)
	if err != nil {
		return err
	}

	c.Manager.BroadcastToRoom(c.RoomID, data)
	return nil
}

func (c *Client) handleJoinRoom(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var join JoinRoomPayload
	if err := json.Unmarshal(data, &join); err != nil {
		return err
	}

	if err := c.Manager.JoinRoom(c, join.RoomID); err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}

	var messages []Message
	err = c.Manager.db.
		Where("room_id = ? AND deleted_at IS NULL", join.RoomID).
		Order("created_at DESC").
		Limit(50).
		Find(&messages).Error
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	history := make([]NewMessagePayload, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		history[len(messages)-1-i] = NewMessagePayload{
			ID:        messages[i].ID,
			RoomID:    messages[i].RoomID,
			SenderID:  messages[i].SenderID,
			Content:   messages[i].Content,
			ReplyToID: messages[i].ReplyToID,
			CreatedAt: messages[i].CreatedAt,
		}
	}

	response := Event{
		Type:    "history",
		Payload: HistoryPayload{Messages: history},
	}

	data, err = json.Marshal(response)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
	default:
		c.Manager.logger.Warn("failed to send history, buffer full", "userID", c.UserID)
	}

	return nil
}

func (c *Client) handleTyping() error {
	if c.RoomID == 0 {
		return errors.New("must join a room first")
	}

	c.Manager.SetTyping(c.RoomID, c.UserID)
	
	typingUsers := c.Manager.GetTypingUsers(c.RoomID)
	
	event := Event{
		Type: "typing",
		Payload: TypingPayload{
			UserIDs: typingUsers,
		},
	}
	
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	
	c.Manager.BroadcastToRoom(c.RoomID, data)
	return nil
}

func (c *Client) sendError(message string) {
	errorEvent := Event{
		Type:    "error",
		Payload: map[string]string{"message": message},
	}
	data, _ := json.Marshal(errorEvent)

	select {
	case c.Send <- data:
	default:
	}
}
