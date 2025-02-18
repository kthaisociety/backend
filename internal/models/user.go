package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email     string  `gorm:"uniqueIndex;not null"`
	Password  *string // Nullable for OAuth users
	Provider  string  `gorm:"not null;default:'credentials'"` // 'credentials' or 'google'
	Profile   Profile
}

