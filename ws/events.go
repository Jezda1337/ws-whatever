package ws

type Event struct {
	Type    string `json:"type"` // voice, send_message, new_message
	From    string `json:"from"`
	Payload string `json:"payload"`
}
