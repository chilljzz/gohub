package model

import "time"

type Message struct {
	ID uint `gorm:"primaryKey;index:idx_conversation_message_cursor,priority:2"`

	ConversationID uint `gorm:"not null;index:idx_conversation_message_cursor,priority:1"`

	SenderID uint `gorm:"not null;index;uniqueIndex:uk_sender_client_message_v2,priority:1"`

	ClientMessageID string `gorm:"size:64;not null;uniqueIndex:uk_sender_client_message_v2,priority:2"`

	Content string `gorm:"type:text;not null"`

	CreatedAt time.Time

	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`

	Sender User `gorm:"foreignKey:SenderID" json:"-"`
}
