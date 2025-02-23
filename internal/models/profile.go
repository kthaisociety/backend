package models

import (
	"gorm.io/gorm"
)

type Profile struct {
	gorm.Model
	UserID         uint   `gorm:"not null;unique"`
	Email          string `gorm:"uniqueIndex;not null" json:"email"`
	FirstName      string `gorm:"not null"`
	LastName       string `gorm:"not null"`
	Image          string `json:"image,omitempty"`
	University     string `json:"university,omitempty"`
	Programme      string `json:"programme,omitempty"`
	GraduationYear int    `json:"graduationYear,omitempty"`
}
