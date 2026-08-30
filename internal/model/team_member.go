package model

import "time"

const (
	TeamRoleOwner  = "owner"
	TeamRoleAdmin  = "admin"
	TeamRoleMember = "member"
)

type TeamMember struct {
	ID     uint   `gorm:"primaryKey"`
	TeamID uint   `gorm:"not null;index;uniqueIndex:uk_team_user"`
	UserID uint   `gorm:"not null;index;uniqueIndex:uk_team_user"`
	Role   string `gorm:"type:varchar(20);not null;default:member"`

	Team Team `gorm:"foreignKey:TeamID" json:"-"`
	User User `gorm:"foreignKey:UserID" json:"-"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
