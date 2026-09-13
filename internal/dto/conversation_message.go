package dto

import "time"

type ConversationMessageResult struct {
	ID uint `json:"id"`

	ConversationID uint `json:"conversation_id"`

	SenderID uint `json:"sender_id"`

	ClientMessageID string `json:"client_message_id"`

	Content string `json:"content"`

	CreatedAt time.Time `json:"created_at"`
}

type ConversationMessagePageResult struct {
	Messages []ConversationMessageResult `json:"messages"`

	NextBeforeID uint `json:"next_before_id,omitempty"`

	HasMore bool `json:"has_more"`
}

type ConversationMessageSyncResult struct {
	Messages    []ConversationMessageResult `json:"messages"`
	NextAfterID uint                        `json:"next_after_id"`
	HasMore     bool                        `json:"has_more"`
}
