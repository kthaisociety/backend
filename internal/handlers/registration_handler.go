package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegistrationHandler struct {
	db *gorm.DB
}

func NewRegistrationHandler(db *gorm.DB) *RegistrationHandler {
	return &RegistrationHandler{db: db}
}

func (h *RegistrationHandler) Register(r *gin.RouterGroup) {
	registrations := r.Group("/registrations")
	{
		// Public endpoints (require auth)
		registrations.Use(middleware.AuthRequired())
		registrations.Use(middleware.RegisteredUserRequired(h.db))
		registrations.GET("", h.List)
		registrations.POST("", h.Create)
		registrations.GET("/:id", h.Get)
		registrations.GET("/my", h.GetUserRegistrations)
		registrations.GET("/event/:eventId", h.GetEventRegistrations)
		registrations.POST("/register/:eventId", h.RegisterForEvent)
		registrations.PUT("/:id/cancel", h.CancelRegistration)

		// Admin-only endpoints
		admin := registrations.Group("/admin")
		admin.Use(middleware.AdminRequired()) // You'll need to create this middleware
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)
		admin.PUT("/:id/status", h.UpdateStatus)
		admin.PUT("/:id/attended", h.MarkAttendance)
	}
}

func (h *RegistrationHandler) List(c *gin.Context) {
	var registrations []models.Registration
	if err := h.db.Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, registrations)
}

func (h *RegistrationHandler) Create(c *gin.Context) {
	var registration models.Registration
	if err := c.ShouldBindJSON(&registration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, registration)
}

func (h *RegistrationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var registration models.Registration
	if err := h.db.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}
	c.JSON(http.StatusOK, registration)
}

func (h *RegistrationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var registration models.Registration
	if err := h.db.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	if err := c.ShouldBindJSON(&registration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, registration)
}

func (h *RegistrationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.Registration{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Registration deleted"})
}

type RegistrationRequest struct {
	EventID             uint                  `json:"event_id" binding:"required"`
	DietaryRestrictions string                `json:"dietary_restrictions"`
	Answers             []CustomAnswerRequest `json:"answers"`
}

type CustomAnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

func (h *RegistrationHandler) RegisterForEvent(c *gin.Context) {
	eventID := c.Param("eventId")
	userID, _, err := h.getUserData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	// Check if event exists
	var event models.Event
	if err := h.db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if registration exists already
	var existingReg models.Registration
	result := h.db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&existingReg)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You are already registered for this event"})
		return
	}

	// Check if event has reached max capacity
	var count int64
	h.db.Model(&models.Registration{}).Where("event_id = ? AND status != ?",
		eventID, models.RegistrationStatusRejected).Count(&count)

	if event.RegistrationMax > 0 && int(count) >= event.RegistrationMax {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event has reached maximum capacity"})
		return
	}

	// Parse request body for custom answers
	var req RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create new registration
	registration := models.Registration{
		EventID:  uint(event.ID),
		UserID:   userID,
		Status:   models.RegistrationStatusPending, // Default to pending
		Attended: false,
	}

	// If event doesn't require approval, auto-approve the registration
	if event.RequiresApproval == nil || !*event.RequiresApproval {
		registration.Status = models.RegistrationStatusApproved
	}

	// Start transaction for registration and answers
	tx := h.db.Begin()

	// Create registration
	if err := tx.Create(&registration).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create custom answers if provided
	if len(req.Answers) > 0 {
		for _, answerReq := range req.Answers {
			answer := models.CustomAnswer{
				RegistrationID: registration.ID,
				QuestionID:     answerReq.QuestionID,
				Answer:         answerReq.Answer,
			}

			if err := tx.Create(&answer).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save custom answers"})
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete registration"})
		return
	}

	c.JSON(http.StatusCreated, registration)
}

// GetUserRegistrations returns all registrations for the current user
func (h *RegistrationHandler) GetUserRegistrations(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Optionally get user profile data
	var profile models.Profile
	h.db.Where("user_id = ?", userID).First(&profile)

	var registrations []models.Registration
	if err := h.db.Where("user_id = ?", userID).
		Preload("Event").Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"registrations": registrations,
		"user": gin.H{
			"id":    userID,
			"email": profile.Email,
			"name":  profile.FirstName + " " + profile.LastName,
		},
	})
}

// GetEventRegistrations returns all registrations for a specific event
func (h *RegistrationHandler) GetEventRegistrations(c *gin.Context) {
	eventID := c.Param("eventId")

	var registrations []models.Registration
	if err := h.db.Where("event_id = ?", eventID).
		Preload("User").Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, registrations)
}

// CancelRegistration allows a user to cancel their own registration
func (h *RegistrationHandler) CancelRegistration(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var registration models.Registration
	if err := h.db.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	// Only allow users to cancel their own registrations
	if registration.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only cancel your own registrations"})
		return
	}

	// Cancel by setting status to rejected
	registration.Status = models.RegistrationStatusRejected

	if err := h.db.Save(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration cancelled successfully"})
}

// UpdateStatus allows admins to update the status of a registration
func (h *RegistrationHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Status models.RegistrationStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var registration models.Registration
	if err := h.db.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	registration.Status = input.Status

	if err := h.db.Save(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, registration)
}

// MarkAttendance allows admins to mark attendance for a registration
func (h *RegistrationHandler) MarkAttendance(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Attended bool `json:"attended" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var registration models.Registration
	if err := h.db.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	registration.Attended = input.Attended

	if err := h.db.Save(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, registration)
}

// getUserData retrieves user data from the session and database
func (h *RegistrationHandler) getUserData(c *gin.Context) (uint, *models.User, error) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return 0, nil, err
	}

	return userID, &user, nil
}
