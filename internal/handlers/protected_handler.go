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
func NewProtectedHandler(db *gorm.DB) *ProtectedHandler {  // Return concrete type
	return &ProtectedHandler{DB: db}
}

func (h *ProtectedHandler) Register(r *gin.RouterGroup) {
	protected := r.Group("")
	{
		protected.GET("", h.Protected)
		protected.GET("/profile", h.Profile)
	}
}

func (h *ProtectedHandler) Protected(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	email := session.Get("email")

	c.JSON(http.StatusOK, gin.H{
		"message": "You have access to protected route",
		"user_id": userID,
		"email":   email,
	})
}

func (h *ProtectedHandler) Profile(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var user struct {
		Email   string         `json:"email"`
		Profile models.Profile `json:"profile"`
	}

	if err := h.DB.Table("users").
		Select("users.email, profiles.*").
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Where("users.id = ?", userID).
		First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	c.JSON(http.StatusOK, user)
} 