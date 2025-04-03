package models

type Startup struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"not null" json:"description"`
	Image       string `gorm:"not null" json:"image"`
	Year        string `gorm:"not null" json:"year"`
}
