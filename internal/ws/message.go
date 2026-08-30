package ws

type IncomingMessage struct {
	Type      string `json:"type"`
	ChannelID uint   `json:"channel_id"`
	Content   string `json:"content"`
}

type OutgoingMessage struct {
	Type      string `json:"type"`
	MessageID uint   `json:"message_id,omitempty"`
	ChannelID uint   `json:"channel_id,omitempty"`
	UserID    uint   `json:"user_id,omitempty"`
	Content   string `json:"content,omitempty"`
	Message   string `json:"message,omitempty"`
	SentAt    string `json:"sent_at,omitempty"`
}

type MessageHandler interface {
	Handle(client *Client, data []byte)
}
