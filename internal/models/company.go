package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Company struct {
	gorm.Model
	Id          uuid.UUID `gorm:"uniqueIndex" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Logo        uuid.UUID `json:"logo"` // reference to a blob_data object
}
