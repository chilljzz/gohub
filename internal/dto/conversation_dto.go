package dto

import "time"

type DirectConversationResult struct {
	ConversationID uint      `json:"conversation_id"`
	Type           string    `json:"type"`
	OtherUserID    uint      `json:"other_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}
