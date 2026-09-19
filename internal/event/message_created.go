package event

import "time"

const (
	TypeMessageCreated = "message.created"
	MessageCreatedV1   = 1
)

type MessageCreated struct {
	EventID string `json:"event_id"`

	Type string `json:"type"`

	Version int `json:"version"`

	OccurredAt time.Time `json:"occurred_at"`

	MessageID uint `json:"message_id"`

	ConversationID uint `json:"conversation_id"`

	SenderID uint `json:"sender_id"`

	ClientMessageID string `json:"client_message_id"`

	Content string `json:"content"`

	CreatedAt time.Time `json:"created_at"`
}
