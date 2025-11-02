package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudyProgram string

const (
	StudyProgramMachineLearning                 StudyProgram = "Machine Learning"
	StudyProgramAppliedMathematics              StudyProgram = "Applied Mathematics"
	StudyProgramBioTechnology                   StudyProgram = "Bio Technology"
	StudyProgramEngineeringPhysics              StudyProgram = "Engineering Physics"
	StudyProgramComputerScience                 StudyProgram = "Computer Science"
	StudyProgramElectricalEngineering           StudyProgram = "Electrical Engineering"
	StudyProgramIndustrialManagement            StudyProgram = "Industrial Management"
	StudyProgramInformationAndCommunicationTech StudyProgram = "Information and Communication Technology"
	StudyProgramChemicalScienceAndEngineering   StudyProgram = "Chemical Science and Engineering"
	StudyProgramMechanicalEngineering           StudyProgram = "Mechanical Engineering"
	StudyProgramMathematics                     StudyProgram = "Mathematics"
	StudyProgramMaterialScienceAndEngineering   StudyProgram = "Material Science and Engineering"
	StudyProgramMedicalEngineering              StudyProgram = "Medical Engineering"
	StudyProgramEnvironmentalEngineering        StudyProgram = "Environmental Engineering"
	StudyProgramTheBuiltEnvironment             StudyProgram = "The Built Environment"
	StudyProgramTechnologyAndEconomics          StudyProgram = "Technology and Economics"
	StudyProgramTechnologyAndHealth             StudyProgram = "Technology and Health"
	StudyProgramTechnologyAndLearning           StudyProgram = "Technology and Learning"
	StudyProgramTechnologyAndManagement         StudyProgram = "Technology and Management"
)

// viv - might need to update with createdAt/updatedAt to track activity and if registerd
type Profile struct {
	gorm.Model
	UserID         uuid.UUID    `gorm:"not null;unique" json:"user_id"`
	Email          string       `gorm:"uniqueIndex;not null" json:"email"`
	FirstName      string       `gorm:"not null" json:"first_name"`
	LastName       string       `gorm:"not null" json:"last_name"`
	University     string       `gorm:"not null" json:"university"`
	Programme      StudyProgram `gorm:"not null" json:"programme,omitempty"`
	GraduationYear int          `gorm:"not null" json:"graduation_year,omitempty"`
	GitHubLink     string       `json:"github_link,omitempty"`
	LinkedInLink   string       `json:"linkedin_link,omitempty"`
	Registered     bool         `json:"registered"`
}
