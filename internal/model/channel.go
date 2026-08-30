package model

import "time"

type Channel struct {
	ID          uint   `gorm:"primaryKey"`
	TeamID      uint   `gorm:"not null;index;uniqueIndex:uk_team_channel_name"`
	Name        string `gorm:"type:varchar(50);not null;uniqueIndex:uk_team_channl_name"`
	Description string `gorm:"type:varchar(200)"`
	CreatedBy   uint   `gorm:"not null;index"`

	Team    Team `gorm:"foreignKey:TeamID" json:"-"`
	Creator User `gorm:"foreignKey:CreatedBy" json:"-"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
