package model

import "time"

const (
	FriendRequestPending  = "pending"
	FriendRequestAccepted = "accepted"
	FriendRequestRejected = "rejected"
)

type FriendRequest struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	SenderID   uint `gorm:"not null;index;uniqueIndex:uk_sender_receiver" json:"sender_id"`
	ReceiverID uint `gorm:"not null;index;uniqueIndex:uk_sender_receiver" json:"receiver_id"`

	Status string `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`

	Sender   User `gorm:"foreignKey:SenderID" json:"-"`
	Receiver User `gorm:"foreignKey:ReceiverID" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Friendship struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID   uint `gorm:"not null;index;uniqueIndex:uk_user_friend" json:"user_id"`
	FriendID uint `gorm:"not null;index;uniqueIndex:uk_user_friend" json:"friedn_id"`

	CreatedAt time.Time `json:"created_at"`
}
