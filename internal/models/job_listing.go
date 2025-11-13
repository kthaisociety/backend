package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobListing struct {
	gorm.Model
	Id              uuid.UUID `gorm:"uniqueIndex" json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Salary          string    `json:"salary"` // usually a range, or list of ints?
	Location        string    `json:"location"`
	StartDate       time.Time `json:"startdate"`
	ApplicationDate time.Time `json:"appdate"`
	PostedDate      time.Time `json:"posteddate"`
	CompanyId       uuid.UUID `json:"companyid"`
}
