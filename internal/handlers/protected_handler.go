package handlers

import (
	"net/http"

	"backend/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Make sure ProtectedHandler implements Handler interface
var _ Handler = (*ProtectedHandler)(nil)

// Export the type by making it public
type ProtectedHandler struct {
	DB *gorm.DB
}

// Export the constructor
func NewProtectedHandler(db *gorm.DB) Handler {
	return &ProtectedHandler{DB: db}
}

func (h *ProtectedHandler) Register(r *gin.RouterGroup) {
	r.GET("/me", h.GetMe)
	r.PUT("/me", h.UpdateMe)
}

func (h *ProtectedHandler) GetMe(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var profile models.Profile
	if err := h.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": profile,
	})
}

func (h *ProtectedHandler) UpdateMe(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var updateData struct {
		Email          string `json:"email"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		Image          string `json:"image"`
		University     string `json:"university"`
		Programme      string `json:"programme"`
		GraduationYear int    `json:"graduationYear"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the profile
	var profile models.Profile
	if err := h.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	// Update profile fields
	profile.Email = updateData.Email
	profile.FirstName = updateData.FirstName
	profile.LastName = updateData.LastName
	profile.Image = updateData.Image
	profile.University = updateData.University
	profile.Programme = updateData.Programme
	profile.GraduationYear = updateData.GraduationYear

	if err := h.DB.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    profile,
	})
}
