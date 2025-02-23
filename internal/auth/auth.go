package auth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"backend/internal/models"

	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Add this line to ensure AuthHandler implements Handler interface
var _ handlers.Handler = (*AuthHandler)(nil)

// Add these constants at the top of the file
const (
	bcryptCost = 12 // Higher than default (10), but not too slow
	sessionName = "kthais_session"
)

func InitAuth() error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	fmt.Printf("InitAuth - Client ID length: %d\n", len(clientID))
	fmt.Printf("InitAuth - Client Secret length: %d\n", len(clientSecret))

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("google oauth credentials not configured")
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	goth.UseProviders(
		google.New(
			clientID,
			clientSecret,
			cfg.BackendURL + "/api/v1/auth/google/callback",
			"email",             // Minimal scope
			"profile",           // For user info
			"openid",           // Enable OpenID Connect
			"https://www.googleapis.com/auth/userinfo.profile", // Explicit profile access
		),
	)

	// Configure provider options
	provider, err := goth.GetProvider("google")
	if err != nil {
		return fmt.Errorf("failed to get google provider: %v", err)
	}

	if googleProvider, ok := provider.(*google.Provider); ok {
		googleProvider.SetHostedDomain("") // Optional: restrict to specific domain
		googleProvider.SetPrompt("select_account consent") // Force consent screen
	}

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
		// Apply rate limiting to OAuth routes
		oauth := auth.Group("/")
		oauth.Use(middleware.RateLimit())
		{
			oauth.GET("/google", h.BeginGoogleAuth)
			oauth.GET("/google/callback", h.GoogleCallback)
		}
		
		// Credential auth routes
		auth.POST("/register", h.RegisterUser)
		auth.POST("/login", h.LoginUser)
		auth.GET("/logout", h.Logout)
		auth.GET("/status", h.Status)
		auth.GET("/user", h.GetUser)
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

// Update the RegisterRequest struct to match frontend fields
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
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

	// Get the auth URL
	state := "random-string" // In production, use a secure random string
	authURL, err := provider.BeginAuth(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin auth"})
		return
	}

	url, err := authURL.GetAuthURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get auth URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if !h.enabled {
		redirectWithError(c, "OAuth is not configured")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	provider, err := goth.GetProvider("google")
	if err != nil {
		log.Printf("Failed to get provider: %v", err)
		redirectWithError(c, "Authentication failed")
		return
	}

	params := c.Request.URL.Query()
	state := params.Get("state")
	gothSession, err := provider.BeginAuth(state)
	if err != nil {
		log.Printf("Failed to begin auth: %v", err)
		redirectWithError(c, "Authentication failed")
		return
	}

	_, err = gothSession.Authorize(provider, params)
	if err != nil {
		log.Printf("Failed to authorize: %v", err)
		redirectWithError(c, fmt.Sprintf("Failed to authorize: %v", err))
		return
	}

	gothUser, err := provider.FetchUser(gothSession)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		redirectWithError(c, "Failed to fetch user data")
		return
	}

	log.Printf("Google User Data: %+v", gothUser)
	log.Printf("Raw Data: %+v", gothUser.RawData)

	// Extract name from RawData
	var firstName, lastName string
	if given, ok := gothUser.RawData["given_name"].(string); ok {
		firstName = given
	}
	if family, ok := gothUser.RawData["family_name"].(string); ok {
		lastName = family
	}

	// If given_name/family_name not found, try to parse from Name
	if firstName == "" || lastName == "" && gothUser.Name != "" {
		names := strings.Split(gothUser.Name, " ")
		if len(names) >= 2 {
			if firstName == "" {
				firstName = names[0]
			}
			if lastName == "" {
				lastName = strings.Join(names[1:], " ")
			}
		} else if len(names) == 1 {
			if firstName == "" {
				firstName = names[0]
			}
		}
	}

	// Check if user exists
	var user models.User
	result := h.db.Where("email = ?", gothUser.Email).First(&user)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:     gothUser.Email,
			Provider:  "google",
			FirstName: firstName,
			LastName:  lastName,
			Image:     gothUser.AvatarURL,
		}

		if err := h.db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			redirectWithError(c, "Failed to create account")
			return
		}
	} else if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		redirectWithError(c, "Database error")
		return
	} else {
		// Update existing user
		user.Provider = "google"
		user.FirstName = firstName
		user.LastName = lastName
		user.Image = gothUser.AvatarURL

		if err := h.db.Save(&user).Error; err != nil {
			log.Printf("Failed to update user: %v", err)
			redirectWithError(c, "Failed to update account")
			return
		}
	}

	// Set session with explicit domain
	session := sessions.Default(c)
	session.Clear()
	session.Set("user_id", user.ID)
	session.Set("email", user.Email)
	session.Set("authenticated", true) // Add explicit authentication flag
	
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %v", err)
		redirectWithError(c, "Failed to create session")
		return
	}

	// Redirect to frontend with explicit success parameter
	frontendURL := cfg.FrontendURL 
	
	dashboardURL := fmt.Sprintf("%s/dashboard?auth=success", frontendURL)
	c.Redirect(http.StatusTemporaryRedirect, dashboardURL)
}

// Helper function to redirect with error
func redirectWithError(c *gin.Context, message string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	frontendURL := cfg.FrontendURL 

	// URL encode the error message
	encodedError := url.QueryEscape(message)
	redirectURL := fmt.Sprintf("%s/auth/login?error=%s", frontendURL, encodedError)
	
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
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

	// Use stronger hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	hashedPasswordStr := string(hashedPassword)
	user := models.User{
		Email:     req.Email,
		Password:  &hashedPasswordStr,
		Provider:  "credentials",
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
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
		"user": user,
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
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"provider": user.Provider,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"image":     user.Image,
		},
	})
}

func (h *AuthHandler) Status(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	authenticated := session.Get("authenticated")

	log.Printf("Session check - UserID: %v, Authenticated: %v", userID, authenticated)

	c.JSON(http.StatusOK, gin.H{
		"enabled": h.enabled,
		"authenticated": authenticated != nil && authenticated.(bool),
		"user_id": userID,
		"email": session.Get("email"),
	})
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		log.Printf("Failed to fetch user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		return
	}
	fmt.Println(user)
	// Transform response to maintain API compatibility
	response := gin.H{
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"provider": user.Provider,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"image":     user.Image,
		},
	}

	c.JSON(http.StatusOK, response)
}
