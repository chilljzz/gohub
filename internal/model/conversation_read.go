package model

import "time"

type ConversationRead struct {
	ConversationID uint `gorm:"primaryKey"`

	UserID uint `gorm:"primaryKey;index"`

	LastDeliveredMessageID uint `gorm:"not null;default:0"`

	LastReadMessageID uint `gorm:"not null;default:0"`

	CreatedAt time.Time

	UpdatedAt time.Time
}
