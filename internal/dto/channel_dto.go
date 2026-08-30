package dto

import "time"

type ChannelResult struct {
	ID          uint      `json:"id"`
	TeamID      uint      `json:"team_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}
