package models

type CategoryType string

const (
	CategoryTypeIndustry    CategoryType = "industry"
	CategoryTypeAcademic    CategoryType = "research"
	CategoryTypeStudent     CategoryType = "student"
	CategoryTypeBoardMember CategoryType = "board member"
)

type Speaker struct {
	ID       uint         `gorm:"primarykey" json:"id"`
	Name     string       `gorm:"not null" json:"name"`
	Title    string       `gorm:"not null" json:"title"`
	Cateogry CategoryType `gorm:"not null" json:"category"`
	Events   []Event      `gorm:"many2many:speaker_events;" json:"events"`
}
