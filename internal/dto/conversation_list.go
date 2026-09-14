package dto

import "time"

type ConversationPeerResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type ConversationChannelResult struct {
	ID uint `json:"id"`

	Name string `json:"name"`

	TeamID uint `json:"team_id"`

	TeamName string `json:"team_name"`
}

type ConversationLastMessageResult struct {
	ID uint `json:"id"`

	SenderID uint `json:"sender_id"`

	Content string `json:"content"`

	CreatedAt time.Time `json:"created_at"`
}

type ConversationListItem struct {
	ID uint `json:"id"`

	Type string `json:"type"`

	Peer *ConversationPeerResult `json:"peer,omitempty"`

	Channel *ConversationChannelResult `json:"channel,omitempty"`

	LastMessage *ConversationLastMessageResult `json:"last_message,omitempty"`

	UnreadCount int64 `json:"unread_count"`
}

type ConversationListResult struct {
	Items []ConversationListItem `json:"items"`

	NextCursor string `json:"next_cursor,omitempty"`

	Hasmore bool `json:"has_more"`
}
