package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email     string  `gorm:"uniqueIndex;not null" json:"email"`
	Password  *string `json:"-"` // Nullable for OAuth users
	Provider  string  `gorm:"not null;default:'credentials'" json:"provider"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Image     string  `json:"image,omitempty"`
}

