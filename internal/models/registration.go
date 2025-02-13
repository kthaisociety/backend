package models

import (
	"time"

	"gorm.io/gorm"
)

type Registration struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	EventID   uint           `gorm:"not null" json:"event_id"`
	Event     Event          `gorm:"foreignKey:EventID" json:"event"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Status    string         `gorm:"not null" json:"status"` // pending, approved, rejected
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
