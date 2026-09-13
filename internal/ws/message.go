package ws

type IncomingMessage struct {
	Type            string `json:"type"`
	ChannelID       uint   `json:"channel_id,omitempty"`
	ConversationID  uint   `json:"conversation_id,omitempty"`
	Conversation    string `json:"conversation,omitempty"`
	ClientMessageID string `json:"client_message_id,omitempty"`
	Content         string `json:"content,omitempty"`
}

type OutgoingMessage struct {
	Type           string `json:"type"`
	MessageID      uint   `json:"message_id,omitempty"`
	ChannelID      uint   `json:"channel_id,omitempty"`
	ConversationID uint   `json:"conversation_id,omitempty"`
	UserID         uint   `json:"user_id,omitempty"`
	Content        string `json:"content,omitempty"`
	Message        string `json:"message,omitempty"`
	SentAt         string `json:"sent_at,omitempty"`
}

type MessageHandler interface {
	Handle(client *Client, data []byte)
}

type MessageAck struct {
	Type            string `json:"type"`
	ClientMessageID string `json:"client_message_id"`
	MessageID       uint   `json:"message_id"`
	ChannelID       uint   `json:"channel_id,omitempty"`
	ConversationID  uint   `json:"conversation_id,omitempty"`
}
