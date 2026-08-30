package model

import "time"

type ChannelMessage struct {
	ID uint `gorm:"primaryKey;index:idx_channel_message_cursor,priority:2"`

	ChannelID uint `gorm:"not null;index:idx_channel_message_cursor,priority:1"`
	SenderID  uint `gorm:"not null;index;uniqueIndex:uk_sender_client_message,priority:1" json:"sender_id"`

	Content string `gorm:"type:text;not null"`

	Channel Channel `gorm:"foreignKey:ChannelID" json:"-"`
	Sender  User    `gorm:"foreignKey:SenderID" json:"-"`

	ClientMessageID string `gorm:"size:64;not null;uniqueIndex:uk_sender_client_message,priority:2" json:"client_message_id"`

	CreatedAt time.Time
}
