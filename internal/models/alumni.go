package models

type Alumni struct {
	ID              uint   `gorm:"primarykey" json:"id"`
	Name            string `gorm:"not null" json:"name"`
	Title           string `gorm:"not null" json:"kthais_title"`
	Period          string `gorm:"not null" json:"period"`
	Description     string `gorm:"not null" json:"description"`
	CurrentPosition string `gorm:"not null" json:"current_position"`
}
