package dto

import "time"

type FriendRequestCreatedResult struct {
	ID        uint            `json:"id"`
	Receiver  UserBridfResult `json:"receiver"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

type FriendRequestResult struct {
	ID        uint            `json:"id"`
	Sender    UserBridfResult `json:"sender"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}
