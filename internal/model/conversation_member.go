package model

import "time"

type ConversationMember struct {
	ConversationID uint `gorm:"primaryKey" json:"conversation_id"`

	UserID uint `gorm:"primaryKey;index:idx_conversation_member_user" json:"user_id"`

	JoinedAt time.Time `json:"joined_at"`
}
