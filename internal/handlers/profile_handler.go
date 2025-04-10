package handlers

import (
	"backend/internal/mailchimp"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	db        *gorm.DB
	mailchimp *mailchimp.MailchimpAPI
}

func NewProfileHandler(db *gorm.DB, mailchimp *mailchimp.MailchimpAPI) *ProfileHandler {
	return &ProfileHandler{db: db, mailchimp: mailchimp}
}

func (h *ProfileHandler) Register(r *gin.RouterGroup) {
	profile := r.Group("/profile")
	{
		// Auth required endpoints
		profile.Use(middleware.AuthRequired())
		profile.GET("/", h.GetMyProfile)
		profile.PUT("/", h.UpdateMyProfile)
		profile.POST("/", h.CreateMyProfile)

		// Admin-only endpoints
		admin := profile.Group("/admin")
		admin.Use(middleware.AdminRequired(h.db))
		admin.GET("", h.ListAllProfiles)
		admin.PUT("/:userId", h.UpdateProfile)
		admin.DELETE("/:userId", h.DeleteProfile)
	}
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
		existingProfile.University = input.University
		existingProfile.Programme = input.Programme
		existingProfile.GraduationYear = input.GraduationYear
		existingProfile.GitHubLink = input.GitHubLink
		existingProfile.LinkedInLink = input.LinkedInLink

		if err := h.db.Save(&existingProfile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update member in Mailchimp
		memberRequest := mailchimp.MemberRequest{
			Email:  existingProfile.Email,
			Status: mailchimp.Subscribed,
			MergeFields: mailchimp.MergeFields{
				FirstName:      existingProfile.FirstName,
				LastName:       existingProfile.LastName,
				Programme:      string(existingProfile.Programme),
				GraduationYear: existingProfile.GraduationYear,
			},
		}
		if _, err := h.mailchimp.UpdateMember(&existingProfile.Email, &memberRequest); err != nil {
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

	// Add member to Mailchimp
	if err := h.mailchimp.SubscribeMember(&newProfile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newProfile)
}

// CreateMyProfile creates a profile for the authenticated user
func (h *ProfileHandler) CreateMyProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Check if profile already exists
	var existingProfile models.Profile
	result := h.db.Where("user_id = ?", userID).First(&existingProfile)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Profile already exists",
			"profile": existingProfile,
		})
		return
	}

	// Parse input
	var input struct {
		FirstName      string              `json:"firstName" binding:"required"`
		LastName       string              `json:"lastName" binding:"required"`
		Email          string              `json:"email" binding:"required,email"`
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

	// Create new profile
	newProfile := models.Profile{
		UserID:         userID,
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Email:          input.Email,
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

	// Add member to Mailchimp
	if err := h.mailchimp.SubscribeMember(&newProfile); err != nil {
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

	// Update member in Mailchimp
	memberRequest := mailchimp.MemberRequest{
		Email:  profile.Email,
		Status: mailchimp.Subscribed,
		MergeFields: mailchimp.MergeFields{
			FirstName:      profile.FirstName,
			LastName:       profile.LastName,
			Programme:      string(profile.Programme),
			GraduationYear: profile.GraduationYear,
		},
	}
	if _, err := h.mailchimp.UpdateMember(&profile.Email, &memberRequest); err != nil {
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
