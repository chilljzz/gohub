package model

import "time"

type Team struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(50);not null"`
	Description string `gorm:"type:varchar(200)"`
	OwnerID     uint   `gorm:"not null;index"`

	Owner     User `gorm:"foreignKey:OwnerID" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
