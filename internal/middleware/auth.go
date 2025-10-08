package middleware

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/utils"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func AuthRequiredJWT(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, cookie := range c.Request.Cookies() {
			if cookie.Name == "jwt" {
				valid, _ := utils.ParseAndVerify(cookie.Value, cfg.JwtSigningKey)
				if valid {
					c.Next()
					return
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					c.Abort()
				}
			}
		}
		// no jwt, not authorized
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
	}
}

func RegisteredUserRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		// Check if profile exists
		var existingProfile models.Profile
		result := db.Where("user_id = ?", userID).First(&existingProfile)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			c.Abort()
			return
		}
		c.Next()
	}
}
