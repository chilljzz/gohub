package model

import "time"

type ConversationMember struct {
	ConversationID uint `gorm:"primaryKey" json:"conversation_id"`

	UserID uint `gorm:"primaryKey" json:"user_id"`

	JoinedAt time.Time `json:"joined_at"`
}
