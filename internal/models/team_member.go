package models

import (
	"time"
)

type TeamMemberRole string

const (
	TeamMemberRoleAdmin  TeamMemberRole = "admin"
	TeamMemberRoleMember TeamMemberRole = "member"
)

type TeamMember struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"user"`
	Title        string         `gorm:"not null" json:"title"`
	Role         TeamMemberRole `gorm:"not null" json:"role"`
	GitHubLink   string         `json:"github_link,omitempty"`
	LinkedInLink string         `json:"linkedin_link,omitempty"`
	AcademicYear int            `gorm:"not null" json:"academic_year"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
