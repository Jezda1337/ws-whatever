package domain

type Message struct {
	ID      *string `json:"id"`
	Type    string  `json:"type"`
	Payload string  `json:"payload"`
}
