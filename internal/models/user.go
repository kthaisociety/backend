package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email     string  `gorm:"uniqueIndex;not null" json:"email"`
	Provider  string  `gorm:"not null;default:'magic-link'" json:"provider"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Image     string  `json:"image,omitempty"`
}

