package ws

import "time"

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type SendMessagePayload struct {
	Content   string `json:"content"`
	ReplyToID *int   `json:"reply_to_id,omitempty"`
}

type NewMessagePayload struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	ReplyToID *int      `json:"reply_to_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type JoinRoomPayload struct {
	RoomID int `json:"room_id"`
}

type HistoryPayload struct {
	Messages []NewMessagePayload `json:"messages"`
}

type TypingPayload struct {
	UserIDs []int `json:"user_ids"`
}
