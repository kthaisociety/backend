package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	db *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

func (h *ProfileHandler) Register(r *gin.RouterGroup) {
	profiles := r.Group("/profiles")
	{
		// Public endpoints
		profiles.GET("/public/:userId", h.GetPublicProfile)

		// Auth required endpoints
		profiles.Use(middleware.AuthRequired())
		profiles.GET("/", h.GetMyProfile)
		profiles.PUT("/", h.UpdateMyProfile)
		profiles.DELETE("/", h.DeleteMyProfile)

		// Admin-only endpoints
		admin := profiles.Group("/admin")
		admin.Use(middleware.AdminRequired())
		admin.GET("", h.ListAllProfiles)
		admin.PUT("/:userId", h.UpdateProfile)
		admin.DELETE("/:userId", h.DeleteProfile)
	}
}

// GetPublicProfile returns the public profile information for a user
func (h *ProfileHandler) GetPublicProfile(c *gin.Context) {
	userId := c.Param("userId")

	var profile models.Profile
	if err := h.db.Where("user_id = ?", userId).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	// Return only public information
	c.JSON(http.StatusOK, gin.H{
		"firstName":      profile.FirstName,
		"lastName":       profile.LastName,
		"fullName":       profile.FirstName + " " + profile.LastName,
		"image":          profile.Image,
		"university":     profile.University,
		"programme":      profile.Programme,
		"graduationYear": profile.GraduationYear,
		"githubLink":     profile.GitHubLink,
		"linkedInLink":   profile.LinkedInLink,
	})
}

// GetMyProfile returns the current user's profile
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var profile models.Profile
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		// If profile doesn't exist, return empty profile
		c.JSON(http.StatusOK, gin.H{
			"userId": userID,
			"exists": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":         userID,
		"exists":         true,
		"email":          profile.Email,
		"firstName":      profile.FirstName,
		"lastName":       profile.LastName,
		"image":          profile.Image,
		"university":     profile.University,
		"programme":      profile.Programme,
		"graduationYear": profile.GraduationYear,
		"githubLink":     profile.GitHubLink,
		"linkedInLink":   profile.LinkedInLink,
	})
}

// UpdateMyProfile allows a user to update their own profile
func (h *ProfileHandler) UpdateMyProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Check if profile exists
	var existingProfile models.Profile
	result := h.db.Where("user_id = ?", userID).First(&existingProfile)

	// Parse input
	var input struct {
		FirstName      string              `json:"firstName" binding:"required"`
		LastName       string              `json:"lastName" binding:"required"`
		Email          string              `json:"email" binding:"required,email"`
		Image          string              `json:"image"`
		University     string              `json:"university"`
		Programme      models.StudyProgram `json:"programme"`
		GraduationYear int                 `json:"graduationYear"`
		GitHubLink     string              `json:"githubLink"`
		LinkedInLink   string              `json:"linkedinLink"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If profile exists, update it
	if result.Error == nil {
		existingProfile.FirstName = input.FirstName
		existingProfile.LastName = input.LastName
		existingProfile.Email = input.Email
		existingProfile.Image = input.Image
		existingProfile.University = input.University
		existingProfile.Programme = input.Programme
		existingProfile.GraduationYear = input.GraduationYear
		existingProfile.GitHubLink = input.GitHubLink
		existingProfile.LinkedInLink = input.LinkedInLink

		if err := h.db.Save(&existingProfile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, existingProfile)
		return
	}

	// If profile doesn't exist, create it
	newProfile := models.Profile{
		UserID:         userID,
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Email:          input.Email,
		Image:          input.Image,
		University:     input.University,
		Programme:      input.Programme,
		GraduationYear: input.GraduationYear,
		GitHubLink:     input.GitHubLink,
		LinkedInLink:   input.LinkedInLink,
	}

	if err := h.db.Create(&newProfile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newProfile)
}

// GetProfile returns a profile by user ID (requires authentication)
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userId := c.Param("userId")

	var profile models.Profile
	if err := h.db.Where("user_id = ?", userId).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ListAllProfiles returns all profiles (admin only)
func (h *ProfileHandler) ListAllProfiles(c *gin.Context) {
	var profiles []models.Profile
	if err := h.db.Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// UpdateProfile allows an admin to update any profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userId := c.Param("userId")

	var profile models.Profile
	if err := h.db.Where("user_id = ?", userId).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	// Update profile fields
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// DeleteProfile allows an admin to delete a profile
func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
	userId := c.Param("userId")

	// Find profile first to check if it exists
	var profile models.Profile
	if err := h.db.Where("user_id = ?", userId).First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	if err := h.db.Delete(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

func (h *ProfileHandler) DeleteMyProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := h.db.Where("user_id = ?", userID).Delete(&models.Profile{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}
