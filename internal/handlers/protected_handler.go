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

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
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
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Image     string `json:"image"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Update user fields
	user.Email = updateData.Email
	user.FirstName = updateData.FirstName
	user.LastName = updateData.LastName
	user.Image = updateData.Image

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
} 