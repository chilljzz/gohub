package model

import "time"

type ConversationListProjection struct {
	ConversationID uint `gorm:"column:conversation_id"`

	Type ConversationType `gorm:"column:conversation_type"`

	ChannelID *uint `gorm:"column:channel_id"`

	ChannelName *string `gorm:"column:channel_name"`

	TeamID *uint `gorm:"column:team_id"`

	TeamName *string `gorm:"column:team_name"`

	PeerUserID *uint `gorm:"column:peer_user_id"`

	PeerUsername *string `gorm:"column:peer_username"`

	PeerNickname *string `gorm:"column:peer_nickname"`

	PeerAvatar *string `gorm:"column:peer_avatar"`

	LastMessageID *uint `gorm:"column:last_message_id"`

	LastSenderID *uint `gorm:"column:last_sender_id"`

	LastMessageContent *string `gorm:"column:last_message_content"`

	LastMessageAt *time.Time `gorm:"column:last_message_at"`

	LastReadMessageID uint `gorm:"column:last_read_message_id"`

	UnreadCount int64 `gorm:"column:unread_count"`
}
