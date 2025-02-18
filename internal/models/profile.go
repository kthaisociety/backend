package models

import (
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	DisplayName string         `json:"display_name"`
	Bio         string         `json:"bio"`
	Image       string         `json:"image"`
	Location    string         `json:"location"`
	Website     string         `json:"website"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
