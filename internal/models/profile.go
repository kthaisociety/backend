package models

import (
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

type Profile struct {
	gorm.Model
	UserID         uint         `gorm:"not null;unique"`
	Email          string       `gorm:"uniqueIndex;not null" json:"email"`
	FirstName      string       `gorm:"not null"`
	LastName       string       `gorm:"not null"`
	Image          string       `json:"image,omitempty"`
	University     string       `json:"university,omitempty"`
	Programme      StudyProgram `json:"programme,omitempty"`
	GraduationYear int          `json:"graduationYear,omitempty"`
	GitHubLink     string       `json:"githubLink,omitempty"`
	LinkedInLink   string       `json:"linkedinLink,omitempty"`
}
