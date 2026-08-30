package model

import "time"

type ChannelRead struct {
	ID uint `gorm:"primaryKey"`

	ChannelID uint `gorm:"not null;uniqueIndex:uk_channel_user_read"`
	UserID    uint `gorm:"not null;uniqueIndex:uk_channel_user_read"`

	LastReadMessageID uint `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
