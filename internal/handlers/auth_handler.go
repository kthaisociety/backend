package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"backend/internal/mailchimp"
	"backend/internal/models"

	"backend/internal/config"
	"backend/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"gorm.io/gorm"
)

// Add this line to ensure AuthHandler implements Handler interface
type AuthHandler struct {
	db        *gorm.DB
	mailchimp *mailchimp.MailchimpAPI
}

func NewAuthHandler(db *gorm.DB, mailchimp *mailchimp.MailchimpAPI) *AuthHandler {
	return &AuthHandler{db: db, mailchimp: mailchimp}
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

		// Keep only these essential routes
		auth.GET("/logout", h.Logout)
		auth.GET("/authenticated", h.CheckAuth)
	}
}

// Add this helper function at the package level
func isOriginAllowed(origin, allowedOrigin string) bool {
	// If allowedOrigin contains a wildcard
	if strings.Contains(allowedOrigin, "*") {
		// Convert the wildcard pattern to a regex pattern
		// Escape special regex characters and convert * to .*
		pattern := "^" + strings.Replace(
			regexp.QuoteMeta(allowedOrigin),
			"\\*",
			".*",
			-1,
		) + "$"

		matched, err := regexp.MatchString(pattern, origin)
		if err != nil {
			log.Printf("Error matching origin pattern: %v", err)
			return false
		}
		return matched
	}

	// Exact match if no wildcard
	return origin == allowedOrigin
}

func InitAuth(cfg *config.Config) error {
	clientID := cfg.OAuth.GoogleClientID
	clientSecret := cfg.OAuth.GoogleClientSecret

	fmt.Printf("InitAuth - Client ID length: %d\n", len(clientID))
	fmt.Printf("InitAuth - Client Secret length: %d\n", len(clientSecret))

	goth.UseProviders(
		google.New(
			clientID,
			clientSecret,
			cfg.BackendURL+"/api/v1/auth/google/callback",
			"email",   // Minimal scope
			"profile", // For user info
			"openid",  // Enable OpenID Connect
			"https://www.googleapis.com/auth/userinfo.profile", // Explicit profile access
		),
	)

	// Configure provider options
	provider, err := goth.GetProvider("google")
	if err != nil {
		return fmt.Errorf("failed to get google provider: %v", err)
	}

	if googleProvider, ok := provider.(*google.Provider); ok {
		googleProvider.SetHostedDomain("")                 // Optional: restrict to specific domain
		googleProvider.SetPrompt("select_account consent") // Force consent screen
	}

	return nil
}

func (h *AuthHandler) BeginGoogleAuth(c *gin.Context) {
	provider, err := goth.GetProvider("google")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	// Get the origin from the request header
	origin := c.GetHeader("Origin")

	// If Origin header is missing, use the Host header or a default value
	if origin == "" {
		host := c.Request.Host
		// Determine scheme (http/https)
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		origin = fmt.Sprintf("%s://%s", scheme, host)
		log.Printf("Origin header missing, using: %s", origin)
	}

	// Validate that the origin is in the allowed list
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	// Check if origin is allowed using the new helper function
	isAllowed := false
	for _, allowed := range cfg.AllowedOrigins {
		if isOriginAllowed(origin, allowed) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
		return
	}

	// Generate a secure state that includes the origin
	state := fmt.Sprintf("%s|%s", uuid.New().String(), origin)

	// Store the state in the session
	session := sessions.Default(c)
	session.Set("oauth_state", state)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

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
	provider, err := goth.GetProvider("google")
	if err != nil {
		log.Printf("Failed to get provider: %v", err)
		redirectWithError(c, "Authentication failed")
		return
	}

	// Get the state from the query parameters
	params := c.Request.URL.Query()
	receivedState := params.Get("state")

	// Retrieve the stored state from the session
	session := sessions.Default(c)
	expectedState := session.Get("oauth_state")

	// Clear the state from the session immediately
	session.Delete("oauth_state")
	session.Save()

	// Verify the state matches
	if expectedState == nil || receivedState != expectedState.(string) {
		log.Printf("State mismatch: expected %v, got %v", expectedState, receivedState)
		redirectWithError(c, "Invalid authentication state")
		return
	}

	// Extract the frontend URL from the state
	stateParts := strings.Split(receivedState, "|")
	if len(stateParts) != 2 {
		log.Printf("Invalid state format")
		redirectWithError(c, "Invalid authentication state")
		return
	}
	frontendURL := stateParts[1]

	gothSession, err := provider.BeginAuth(receivedState)
	if err != nil {
		log.Printf("Failed to begin auth: %v", err)
		redirectWithError(c, fmt.Sprintf("Failed to authorize: %v", err))
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
			Email:    gothUser.Email,
			Provider: "google",
		}

		if err := h.db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			redirectWithError(c, "Failed to create account")
			return
		}
	} else {
		log.Printf("Database error: %v", result.Error)
		redirectWithError(c, "Database error")
		return
	}

	var profile models.Profile
	if err := h.db.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Set session
			session = sessions.Default(c)
			session.Clear()
			session.Set("user_id", user.ID)
			session.Set("authenticated", true)

			// Redirect to frontend
			dashboardURL := fmt.Sprintf("%s/auth/complete-registration?fname=%s&lname=%s", frontendURL, firstName, lastName)
			c.Redirect(http.StatusTemporaryRedirect, dashboardURL)
		}
	}

	// Set session
	session = sessions.Default(c)
	session.Clear()
	session.Set("user_id", user.ID)
	session.Set("authenticated", true)

	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %v", err)
		redirectWithError(c, "Failed to create session")
		return
	}

	// Redirect to frontend
	dashboardURL := fmt.Sprintf("%s/dashboard?auth=success", frontendURL)
	c.Redirect(http.StatusTemporaryRedirect, dashboardURL)
}

// Update the redirectWithError function to use the frontend URL from state
func redirectWithError(c *gin.Context, message string) {
	// Get the state from the query parameters
	params := c.Request.URL.Query()
	receivedState := params.Get("state")

	// Extract the frontend URL from the state
	stateParts := strings.Split(receivedState, "|")
	if len(stateParts) != 2 {
		log.Printf("Invalid state format in error redirect")
		return
	}
	frontendURL := stateParts[1]

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

func (h *AuthHandler) CheckAuth(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	authenticated := session.Get("authenticated")

	isAuthenticated := authenticated != nil && authenticated.(bool)

	// Prepare response
	response := gin.H{
		"authenticated": isAuthenticated,
		"user_id":       userID,
	}

	// If authenticated, verify user exists in DB and get user details
	if isAuthenticated && userID != nil {
		var user models.User
		result := h.db.First(&user, userID)

		if result.Error == nil {
			// User found, add details to response
			response["email"] = user.Email
			response["provider"] = user.Provider
			response["created_at"] = user.CreatedAt
			response["updated_at"] = user.UpdatedAt
		} else {
			// User not found or DB error
			log.Printf("Error retrieving user %v: %v", userID, result.Error)
			response["authenticated"] = false
			session.Clear()
			session.Save()
		}
	}

	c.JSON(http.StatusOK, response)
}
