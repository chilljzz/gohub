package dto

import (
	"time"
)

type ChannelMessageResult struct {
	ID              uint      `json:"id"`
	ChannelID       uint      `json:"channel_id"`
	SenderID        uint      `json:"sender_id"`
	ClientMessageID string    `json:"client_message_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
}

type ChannelMessagePageResult struct {
	Messages []ChannelMessageResult `json:"messages"`

	NextBeforeID uint `json:"next_before_id"`
	HasMore      bool `json:"has_more"`
}

type ChannelMessageSyncResult struct {
	Messages []ChannelMessageResult `json:"messages"`

	NextAfterID uint `json:"next_after_id"`
	HasMore     bool `json:"has_more"`
}

type ChannelUnreadResult struct {
	ChannelID         uint  `json:"channel_id"`
	LastReadMessageID uint  `json:"last_read_message_id"`
	UnreadCount       int64 `json:"unread_count"`
}
