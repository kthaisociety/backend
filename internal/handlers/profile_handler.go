package handlers

import (
	"net/http"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProfileHandler handles profile-related endpoints.
type ProfileHandler struct {
	db *gorm.DB
}

// New instance of ProfileHandler
func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

// Register sets up the profile endpoints under /profile
func (h *ProfileHandler) Register(r *gin.RouterGroup) {
	profiles := r.Group("/profiles")
	{
		profiles.GET("", h.List)
		profiles.POST("", h.Create)
		profiles.GET("/:id", h.Get)
		profiles.PUT("/:id", h.Update)
		profiles.DELETE("/:id", h.Delete)
	}
}

// List retrieves all profiles.
func (h *ProfileHandler) List(c *gin.Context) {
	var profiles []models.Profile
	if err := h.db.Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profiles)
}

// Create a new profile
func (h *ProfileHandler) Create(c *gin.Context) {
	var profile models.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, profile)
}

// Get retrieves a specific profile by its ID.
func (h *ProfileHandler) Get(c *gin.Context) {
	var profile models.Profile
	if err := h.db.First(&profile, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// Update modifies an existing profile.
func (h *ProfileHandler) Update(c *gin.Context) {
	var profile models.Profile
	if err := h.db.First(&profile, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

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

// Delete removes a profile.
func (h *ProfileHandler) Delete(c *gin.Context) {
	if err := h.db.Delete(&models.Profile{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted"})
}

