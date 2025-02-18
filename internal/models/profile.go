package models

import (
	"gorm.io/gorm"
)

type Profile struct {
	gorm.Model
	UserID    uint   `gorm:"uniqueIndex"`
	FirstName string
	LastName  string
	Image     string
} 