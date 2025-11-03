package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"backend/internal/mailchimp"
	"backend/internal/models"
	"backend/internal/utils"

	"backend/internal/config"
	"backend/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/linkedin"
	"gorm.io/gorm"
)

// Add this line to ensure AuthHandler implements Handler interface
type AuthHandler struct {
	db            *gorm.DB
	mailchimp     *mailchimp.MailchimpAPI
	jwtSigningKey string
}

func NewAuthHandler(db *gorm.DB, mailchimp *mailchimp.MailchimpAPI, skey string) *AuthHandler {
	return &AuthHandler{db: db, mailchimp: mailchimp, jwtSigningKey: skey}
}

// Update Register method to match the Handler interface
func (h *AuthHandler) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		// Apply rate limiting to OAuth routes
		oauth := auth.Group("/")
		oauth.Use(middleware.RateLimit())
		{	//Use a generic  : provider parameter
			oauth.GET("/:provider", h.BeginAuth)
            oauth.GET("/:provider/callback", h.AuthCallback)
		}

		// Keep only these essential routes
		auth.GET("/status", h.Status)
		auth.GET("/refresh_token", h.RefreshToken)
		auth.GET("/logout", h.Logout)
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

// viv - usually a separate refresh token is used but I don't know why that is necessary
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	old_token := utils.GetJWT(c)
	claims := utils.GetClaims(old_token)
	userId, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		log.Printf("Refresh failed\n")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user id"})
		return
	}
	var user models.User
	result := h.db.Where("user_id = ?", userId).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retreive user info"})
	}
	newToken := utils.WriteJWT(user.Email, user.Roles, user.UserId, h.jwtSigningKey, 15)
	c.SetCookie("jwt", newToken, 3600, "/", "localhost:3000", false, false)
}

func InitAuth(cfg *config.Config) error {
	// Google
    goth.UseProviders(
        google.New(
            cfg.OAuth.GoogleClientID,
            cfg.OAuth.GoogleClientSecret,
            cfg.BackendURL+"/api/v1/auth/google/callback", 
            "email",
            "profile",
            "openid",
            "https://www.googleapis.com/auth/userinfo.profile",
        ),
        // LinkedIn
        linkedin.New(
            cfg.OAuth.LinkedInClientID,
            cfg.OAuth.LinkedInClientSecret,
            cfg.BackendURL+"/api/v1/auth/linkedin/callback", // LinkedIn needs its own callback URL
            "email", "profile", "openid", // Standard scopes
        ),
        
    )

    // Configure Google provider options
    provider, err := goth.GetProvider("google")
    if err != nil {
        return fmt.Errorf("failed to get google provider: %v", err)
    }
    if googleProvider, ok := provider.(*google.Provider); ok {
        googleProvider.SetHostedDomain("")
        googleProvider.SetPrompt("select_account consent")
    }

    // Configure LinkedIn provider options (optional, but good to check)
    if cfg.OAuth.LinkedInClientID != "" { // Only if configured
        _, err = goth.GetProvider("linkedin")
        if err != nil {
            return fmt.Errorf("failed to get linkedin provider: %v", err)
        }
    }

    return nil
}

func (h *AuthHandler) Status(c *gin.Context) {
	token_str := utils.GetJWTString(c)
	valid, _ := utils.ParseAndVerify(token_str, h.jwtSigningKey)
	if !valid {
		c.JSON(401, gin.H{"authenticate": false})
	} else {
		c.JSON(200, gin.H{"authenticate": true})
	}

}

func (h *AuthHandler) BeginAuth(c *gin.Context) {
	providerName := c.Param("provider")
	provider, err := goth.GetProvider(providerName)
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

func (h *AuthHandler) AuthCallback(c *gin.Context) {
    providerName := c.Param("provider")

    // --- 1. STATE VALIDATION (This logic was already generic) ---
    provider, err := goth.GetProvider(providerName)
    if err != nil {
        log.Printf("Failed to get provider: %v", err)
        redirectWithError(c, "Authentication failed")
        return
    }

    params := c.Request.URL.Query()
    receivedState := params.Get("state")
    session := sessions.Default(c)
    expectedState := session.Get("oauth_state")

    session.Delete("oauth_state")
    session.Save()

    if expectedState == nil || receivedState != expectedState.(string) {
        log.Printf("State mismatch: expected %v, got %v", expectedState, receivedState)
        redirectWithError(c, "Invalid authentication state")
        return
    }

    stateParts := strings.Split(receivedState, "|")
    if len(stateParts) != 2 {
        log.Printf("Invalid state format")
        redirectWithError(c, "Invalid authentication state")
        return
    }
    frontendURL := stateParts[1]

    // --- 2. GOTH AUTHORIZATION (This logic was also generic) ---
    gothSession, err := provider.BeginAuth(receivedState)
    if err != nil {
        log.Printf("Failed to begin auth: %v", err)
        redirectWithError(c, fmt.Sprintf("Failed to authorize: %v", err))
        return
    }

    _, err = gothSession.Authorize(provider, params)
    if err != nil {
        log.Printf("Failed to authorize: %T %v", err, err)
        redirectWithError(c, fmt.Sprintf("Failed to authorize: %v", err))
        return
    }

    // --- 3. DISPATCH TO GET USER DATA (The new part) ---
    var pUser *providerUser

    switch providerName {
    case "google":
        pUser, err = h.getUserDataFromGoogle(gothSession)
    case "linkedin":
        pUser, err = h.getUserDataFromLinkedIn(gothSession)
    default:
        err = fmt.Errorf("unknown provider: %s", providerName)
    }

    if err != nil {
        log.Printf("Failed to get user data for provider %s: %v", providerName, err)
        redirectWithError(c, "Failed to get user information")
        return
    }

    // --- 4. CALL GENERIC LOGIN/CREATE FUNCTION ---
    h.processOAuthLogin(c, pUser, frontendURL)
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


// providerUser is a standardized struct to hold data from any provider
type providerUser struct {
    Email     string
    FirstName string
    LastName  string
    Provider  string
}

// processOAuthLogin handles the GENERIC logic for finding/creating a user
// and issuing a JWT. This is the logic you wanted to refactor.
func (h *AuthHandler) processOAuthLogin(c *gin.Context, pUser *providerUser, frontendURL string) {
    // --- THIS IS YOUR LOGIC, MOVED FROM GoogleCallback ---

    // Check if user exists
    var user models.User
    result := h.db.Where("email = ?", pUser.Email).First(&user)

    if result.Error == gorm.ErrRecordNotFound {
        // Create new user
        user = models.User{
            Email:     pUser.Email,
            Provider:  pUser.Provider, // Use the provider name
            Roles:     []string{"user"},
            UserId:    uuid.New(),
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        if err := h.db.Create(&user).Error; err != nil {
            log.Printf("Failed to create user: %v", err)
            redirectWithError(c, "Failed to create account")
            return
        }
    } else if result.Error != nil {
        // Database error (not "record not found")
        log.Printf("Database error: %v", result.Error)
        redirectWithError(c, "Database error")
        return
    }
    // If no error, user already exists and was loaded successfully

    // Check if profile exists
    var profile models.Profile
    profileExists := h.db.Where("user_id = ?", user.UserId).First(&profile).Error == nil
    if !profileExists {
        profile.UserID = user.UserId
        profile.Email = user.Email
        profile.FirstName = pUser.FirstName // Use generic data
        profile.LastName = pUser.LastName  // Use generic data
        profile.Registered = false
        if err := h.db.Create(&profile).Error; err != nil {
            log.Printf("Failed to create profile for user: %v\n", profile)
        }
    }

    // Set session for the user
    session := sessions.Default(c)
    session.Clear()
    session.Set("user_id", user.ID)
    session.Set("authenticated", true)

    if err := session.Save(); err != nil {
        log.Printf("Failed to save session: %v", err)
        redirectWithError(c, "Failed to create session")
        return
    }

    // Redirect based on whether profile exists
    var dashboardURL string
    if profileExists && profile.Registered {
        dashboardURL = fmt.Sprintf("%s/dashboard?auth=success", frontendURL)
    } else {
        dashboardURL = fmt.Sprintf("%s/auth/complete-registration?fname=%s&lname=%s", frontendURL, pUser.FirstName, pUser.LastName)
    }

    // Issue our own JWT
    authJwt := utils.WriteJWT(user.Email, user.Roles, user.UserId, h.jwtSigningKey, 15)
    c.SetCookie("jwt", authJwt, 3600, "/", "localhost", false, false)
    c.Redirect(http.StatusTemporaryRedirect, dashboardURL)
}

// getUserDataFromGoogle extracts user info from Google's ID token
// This is the Google-SPECIFIC logic, isolated.
func (h *AuthHandler) getUserDataFromGoogle(gothSession goth.Session) (*providerUser, error) {
    gSession, ok := gothSession.(*google.Session)
    if !ok {
        return nil, fmt.Errorf("failed to cast session to google.Session")
    }
    
    // This is your existing, correct logic for parsing the Google token
    valid, token := utils.ParseAndVerifyGoogle(gSession.IDToken)
    if token == nil {
        return nil, fmt.Errorf("error parsing google jwt: %v", gSession.IDToken)
    }
    if !valid {
        return nil, fmt.Errorf("invalid Google Token")
    }

    token_data := utils.GetClaims(token)
    var firstName, lastName, email, name string
    
    if given, ok := token_data["given_name"].(string); ok {
        firstName = given
    }
    if family, ok := token_data["family_name"].(string); ok {
        lastName = family
    }
    if emejl, ok := token_data["email"].(string); ok {
        email = emejl
    }
    if fullname, ok := token_data["name"].(string); ok {
        name = fullname
    }
    
    // Your smart name-splitting logic
    if (firstName == "" || lastName == "") && name != "" {
        names := strings.Split(name, " ")
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

    if email == "" {
        return nil, fmt.Errorf("no email found in Google token")
    }

    return &providerUser{
        Email:     email,
        FirstName: firstName,
        LastName:  lastName,
        Provider:  "google",
    }, nil
}

// getUserDataFromLinkedIn fetches user info using goth's generic FetchUser
// This is the LinkedIn-SPECIFIC logic.
func (h *AuthHandler) getUserDataFromLinkedIn(gothSession goth.Session) (*providerUser, error) {
    // For LinkedIn, we can use the generic FetchUser method from Goth
    // This requires an extra API call, which goth handles for us.
    // Note: This assumes you requested "email", "profile", "openid" scopes in InitAuth.
    
    provider, _ := goth.GetProvider("linkedin")
    gothUser, err := provider.FetchUser(gothSession)
    if err != nil {
        return nil, err
    }
    
    if gothUser.Email == "" {
         return nil, fmt.Errorf("no email returned from LinkedIn")
    }

    return &providerUser{
        Email:     gothUser.Email,
        FirstName: gothUser.FirstName,
        LastName:  gothUser.LastName,
        Provider:  "linkedin",
    }, nil
}


func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
