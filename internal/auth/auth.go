package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"gorm.io/gorm"
)

func InitAuth() error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	fmt.Printf("InitAuth - Client ID length: %d\n", len(clientID))
	fmt.Printf("InitAuth - Client Secret length: %d\n", len(clientSecret))

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("google oauth credentials not configured")
	}

	goth.UseProviders(
		google.New(
			clientID,
			clientSecret,
			"http://localhost:8080/api/v1/auth/google/callback",
		),
	)
	return nil
}

type AuthHandler struct {
	db *gorm.DB
	enabled bool
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	// Check if OAuth is configured
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	enabled := clientID != "" && clientSecret != ""

	// Add debug logging
	fmt.Printf("Auth Handler Initialization:\n")
	fmt.Printf("Client ID: %s\n", maskString(clientID))
	fmt.Printf("Client Secret: %s\n", maskString(clientSecret))
	fmt.Printf("OAuth enabled: %v\n", enabled)

	return &AuthHandler{
		db:      db,
		enabled: enabled,
	}
}

// Helper function to mask sensitive data
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func (h *AuthHandler) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.GET("/google", h.BeginGoogleAuth)
		auth.GET("/google/callback", h.GoogleCallback)
		auth.GET("/status", func(c *gin.Context) {
			session := sessions.Default(c)
			userID := session.Get("user_id")
			c.JSON(http.StatusOK, gin.H{
				"enabled": h.enabled,
				"authenticated": userID != nil,
				"debug": gin.H{
					"client_id":            maskString(os.Getenv("GOOGLE_CLIENT_ID")),
					"client_secret":        maskString(os.Getenv("GOOGLE_CLIENT_SECRET")),
					"client_id_exists":     os.Getenv("GOOGLE_CLIENT_ID") != "",
					"client_secret_exists": os.Getenv("GOOGLE_CLIENT_SECRET") != "",
					"working_dir":          getWorkingDir(),
					"env_files":            findEnvFiles(),
				},
			})
		})
		auth.GET("/logout", h.Logout)
	}
}

func (h *AuthHandler) BeginGoogleAuth(c *gin.Context) {
	if !h.enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth is not configured"})
		return
	}
	provider, err := goth.GetProvider("google")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	state := "random-string" // In production, use a proper state management
	session, err := provider.BeginAuth(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin auth"})
		return
	}

	url, err := session.GetAuthURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get auth url"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if !h.enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OAuth is not configured"})
		return
	}

	provider, err := goth.GetProvider("google")
	if err != nil {
		log.Printf("Failed to get provider: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	// Add logging for debugging
	log.Printf("Processing callback with params: %+v", c.Request.URL.Query())

	params := c.Request.URL.Query()
	state := params.Get("state")
	gothSession, err := provider.BeginAuth(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin auth"})
		return
	}

	_, err = gothSession.Authorize(provider, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to authorize: %v", err)})
		return
	}

	gothUser, err := provider.FetchUser(gothSession)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Check if user exists
	var user models.User
	result := h.db.Where("email = ?", gothUser.Email).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:    gothUser.Email,
			Provider: "google",
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create user: %v", err)})
			return
		}

		// Create profile with enhanced Google data
		profile := models.Profile{
			UserID:      user.ID,
			FirstName:   gothUser.FirstName,
			LastName:    gothUser.LastName,
			DisplayName: gothUser.Name,
			Image:       gothUser.AvatarURL,
			Website:     gothUser.Location,  // Google might provide location
			Location:    gothUser.Location,
		}

		// If we have additional data from Google, add it
		if gothUser.Description != "" {
			profile.Bio = gothUser.Description
		}

		if err := h.db.Create(&profile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create profile: %v", err)})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", result.Error)})
		return
	} else {
		// User exists, update their profile with any new information
		var profile models.Profile
		if err := h.db.Where("user_id = ?", user.ID).First(&profile).Error; err == nil {
			// Update profile with any new information from Google
			profile.FirstName = gothUser.FirstName
			profile.LastName = gothUser.LastName
			profile.DisplayName = gothUser.Name
			profile.Image = gothUser.AvatarURL
			
			if err := h.db.Save(&profile).Error; err != nil {
				log.Printf("Failed to update profile: %v", err)
			}
		}
	}

	// Now use the cookie session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("email", user.Email)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save session: %v", err)})
		return
	}

	// Redirect to frontend or return session info
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully authenticated",
		"user": user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("error getting working dir: %v", err)
	}
	return dir
}

func findEnvFiles() []string {
	files := []string{}
	entries, err := os.ReadDir(".")
	if err != nil {
		return []string{fmt.Sprintf("error reading dir: %v", err)}
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && (entry.Name() == ".env" || entry.Name() == ".env.local") {
			info, err := entry.Info()
			if err != nil {
				files = append(files, fmt.Sprintf("%s (error getting info: %v)", entry.Name(), err))
				continue
			}
			files = append(files, fmt.Sprintf("%s (size: %d, mode: %v)", entry.Name(), info.Size(), info.Mode()))
		}
	}
	return files
}
