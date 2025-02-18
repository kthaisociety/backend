package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/internal/models"

	"backend/internal/handlers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Add this line to ensure AuthHandler implements Handler interface
var _ handlers.Handler = (*AuthHandler)(nil)

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

// Update Register method to match the Handler interface
func (h *AuthHandler) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		// OAuth routes
		auth.GET("/google", h.BeginGoogleAuth)
		auth.GET("/google/callback", h.GoogleCallback)
		
		// Credential auth routes
		auth.POST("/register", h.RegisterUser)
		auth.POST("/login", h.LoginUser)
		auth.GET("/logout", h.Logout)
		auth.GET("/status", h.Status)
	}
}

// Update constructor to return Handler interface
func NewAuthHandler(db *gorm.DB) handlers.Handler {
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

// Add these structs for request/response handling
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		// Continue with user creation if user not found
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	hashedPasswordStr := string(hashedPassword)
	user := models.User{
		Email:    req.Email,
		Password: &hashedPasswordStr,
		Provider: "credentials",
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create profile
	profile := models.Profile{
		UserID:    user.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Set session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("email", user.Email)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id": user.ID,
			"email": user.Email,
			"profile": profile,
		},
	})
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	if err := h.db.Where("email = ? AND provider = ?", req.Email, "credentials").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is linked to OAuth provider"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Set session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("email", user.Email)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": user,
	})
}

func (h *AuthHandler) Status(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	c.JSON(http.StatusOK, gin.H{
		"enabled": h.enabled,
		"authenticated": userID != nil,
		"user_id": userID,
		"email": session.Get("email"),
	})
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
