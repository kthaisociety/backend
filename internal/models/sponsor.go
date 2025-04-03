package models

type Sponsor struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `gorm:"not null" json:"name"`
	Logo string `gorm:"not null" json:"logo"`
	Link string `gorm:"not null" json:"link"`
}
