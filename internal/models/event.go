package models

import (
	"time"

	"gorm.io/gorm"
)

type EventType string

const (
	EventTypeLecture    EventType = "lecture"
	EventTypeWorkshop   EventType = "webinar"
	EventTypeSeminar    EventType = "hackathon"
	EventTypeConference EventType = "workshop"
	EventTypeJobFair    EventType = "job fair"
	EventTypeOther      EventType = "other"
	// TODO: add more event types
)

type RegistrationMethod string

const (
	RegistrationMethodWebsite RegistrationMethod = "website"
	RegistrationMethodLink    RegistrationMethod = "link"
	// TODO: add more registration methods
)

type Event struct {
	ID                 uint               `gorm:"primarykey" json:"id"`
	Title              string             `gorm:"not null" json:"title"`
	Description        string             `json:"description"`
	RegistrationMethod RegistrationMethod `json:"registration_method"`
	ICSFileEndpoint    string             `json:"ics_file_endpoint"`
	Location           string             `json:"location"`
	Image              string             `json:"image"`
	RegistrationMax    int                `json:"registration_max"`
	TypeOfEvent        EventType          `json:"type_of_event"`
	StartDate          time.Time          `json:"start_date"`
	EndDate            time.Time          `json:"end_date"`
	CreatedBy          uint               `json:"created_by"`
	User               User               `gorm:"foreignKey:CreatedBy" json:"user"`
	RequiresApproval   *bool              `gorm:"not null;default:true" json:"requires_approval"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`
}
