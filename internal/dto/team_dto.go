package dto

import "time"

type TeamResult struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"owner_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type TeamMemberResult struct {
	ID       uint      `json:"id"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
