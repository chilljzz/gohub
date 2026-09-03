package model

import "time"

type ConversationType string

const (
	ConversationTypeChannel ConversationType = "channel"
	ConversationTypeDirect  ConversationType = "direct"
)

type Conversation struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Type ConversationType `gorm:"type:varchar(16);not null;index" json:"type"`

	ChannelID *uint `gorm:"uniqueIndex" json:"channel_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
