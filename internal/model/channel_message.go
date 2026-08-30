package model

import "time"

type ChannelMessage struct {
	ID uint `gorm:"primaryKey;index:idx_channel_message_cursor,priority:2"`

	ChannelID uint `gorm:"not null;index:idx_channel_message_cursor,priority:1"`
	SenderID  uint `gorm:"not null;index"`

	Content string `gorm:"type:text;not null"`

	Channel Channel `gorm:"foreignKey:ChannelID" json:"-"`
	Sender  User    `gorm:"foreignKey:SenderID" json:"-"`

	CreatedAt time.Time
}
