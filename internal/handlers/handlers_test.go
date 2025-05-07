package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.CustomQuestion{},
		&models.Registration{},
		&models.CustomAnswer{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// create a dummy user for registrations
	db.Create(&models.User{Email: "test@example.com"})

	return db
}

func setupTestRouter(t *testing.T) *gin.Engine {
	router := gin.Default()
	// bypass authentication: set a user_id in context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	db := setupTestDB(t)

	// Register event handler routes
	api := router.Group("/api")
	eventHandler := NewEventHandler(db)
	eventHandler.Register(api)
	// Register only the public registration endpoint for testing
	regHandler := NewRegistrationHandler(db)
	api.POST("/registrations/register/:eventId", regHandler.RegisterForEvent)

	return router
}

func performRequest(r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateEvent(t *testing.T) {
	router := setupTestRouter(t)

	// Create test event data
	requiresApproval := false
	eventData := gin.H{
		"title":             "Test Event",
		"description":       "This is a test event",
		"start_date":        time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"end_date":          time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		"location":          "Test Location",
		"registration_max":  50,
		"requires_approval": requiresApproval,
	}

	// Create event through HTTP endpoint
	w := performRequest(router, "POST", "/api/event", eventData)
	assert.Equal(t, http.StatusCreated, w.Code, "Should return 201 Created")

	// Verify the response contains our event
	var createdEvent models.Event
	err := json.Unmarshal(w.Body.Bytes(), &createdEvent)
	assert.NoError(t, err, "Should unmarshal response without error")
	assert.NotZero(t, createdEvent.ID, "Event ID should be set after creation")
	assert.Equal(t, eventData["title"], createdEvent.Title)
	assert.Equal(t, eventData["description"], createdEvent.Description)
	assert.Equal(t, eventData["location"], createdEvent.Location)
	assert.Equal(t, eventData["registration_max"], createdEvent.RegistrationMax)
	assert.Equal(t, eventData["requires_approval"], *createdEvent.RequiresApproval)

	// Verify event exists in database by retrieving it
	w = performRequest(router, "GET", fmt.Sprintf("/api/event/%d", createdEvent.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code, "Should find created event")

	var foundEvent models.Event
	err = json.Unmarshal(w.Body.Bytes(), &foundEvent)
	assert.NoError(t, err, "Should unmarshal found event without error")
	assert.Equal(t, createdEvent.ID, foundEvent.ID, "Should find the same event")
}

func TestCreateEventWithDefaultApproval(t *testing.T) {
	router := setupTestRouter(t)

	// Create test event data without requires_approval field
	eventData := gin.H{
		"title":            "Test Event No Approval Field",
		"description":      "This is a test event without approval field",
		"start_date":       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"end_date":         time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		"location":         "Test Location",
		"registration_max": 50,
	}

	// Create event through HTTP endpoint
	w := performRequest(router, "POST", "/api/event", eventData)
	assert.Equal(t, http.StatusCreated, w.Code, "Should return 201 Created")

	// Verify the response contains our event
	var createdEvent models.Event
	err := json.Unmarshal(w.Body.Bytes(), &createdEvent)
	assert.NoError(t, err, "Should unmarshal response without error")
	assert.NotNil(t, createdEvent.RequiresApproval, "RequiresApproval should not be nil")
	assert.True(t, *createdEvent.RequiresApproval, "Default value for requires_approval should be true")
}

func TestCreateEventWithExplicitNoApproval(t *testing.T) {
	router := setupTestRouter(t)

	// Create test event data with explicit requires_approval = false
	eventData := gin.H{
		"title":             "Test Event No Approval",
		"description":       "This is a test event with no approval required",
		"start_date":        time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"end_date":          time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		"location":          "Test Location",
		"registration_max":  50,
		"requires_approval": false,
	}

	// Create event through HTTP endpoint
	w := performRequest(router, "POST", "/api/event", eventData)
	assert.Equal(t, http.StatusCreated, w.Code, "Should return 201 Created")

	// Verify the response contains our event
	var createdEvent models.Event
	err := json.Unmarshal(w.Body.Bytes(), &createdEvent)
	assert.NoError(t, err, "Should unmarshal response without error")
	assert.NotNil(t, createdEvent.RequiresApproval, "RequiresApproval should not be nil")
	assert.False(t, *createdEvent.RequiresApproval, "requires_approval should be false")
	log.Println("createdEvent", createdEvent)
}

func TestCreateEventWithCustomQuestions(t *testing.T) {
	router := setupTestRouter(t)

	// Create test event data with custom questions
	eventData := gin.H{
		"title":             "Test Event With Questions",
		"description":       "This is a test event with custom questions",
		"start_date":        time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"end_date":          time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		"location":          "Test Location",
		"registration_max":  50,
		"requires_approval": false,
		"custom_questions": []gin.H{
			{
				"question": "What is your favorite color?",
				"type":     "text",
				"required": true,
			},
			{
				"question": "How many years of experience do you have?",
				"type":     "number",
				"required": false,
			},
			{
				"question": "Select your preferred time slot",
				"type":     "select",
				"required": true,
			},
		},
	}

	// Create event through HTTP endpoint
	w := performRequest(router, "POST", "/api/event", eventData)
	assert.Equal(t, http.StatusCreated, w.Code, "Should return 201 Created")

	// Verify the response contains our event with custom questions
	var createdEvent models.Event
	err := json.Unmarshal(w.Body.Bytes(), &createdEvent)
	assert.NoError(t, err, "Should unmarshal response without error")
	assert.NotZero(t, createdEvent.ID, "Event ID should be set after creation")
	assert.Len(t, createdEvent.CustomQuestions, 3, "Should have 3 custom questions")

	// Verify each custom question
	questions := eventData["custom_questions"].([]gin.H)
	for i, q := range questions {
		assert.Equal(t, q["question"], createdEvent.CustomQuestions[i].Question)
		assert.Equal(t, q["type"], string(createdEvent.CustomQuestions[i].Type))
		assert.Equal(t, q["required"], createdEvent.CustomQuestions[i].Required)
	}

	// Verify we can retrieve the event with its questions
	w = performRequest(router, "GET", fmt.Sprintf("/api/event/%d", createdEvent.ID), nil)
	assert.Equal(t, http.StatusOK, w.Code, "Should find created event")

	var foundEvent models.Event
	err = json.Unmarshal(w.Body.Bytes(), &foundEvent)
	assert.NoError(t, err, "Should unmarshal found event without error")
	assert.Equal(t, createdEvent.ID, foundEvent.ID, "Should find the same event")
	assert.Len(t, foundEvent.CustomQuestions, 3, "Retrieved event should have 3 custom questions")
	// Inverted assertion: ensure the length is not 4
	assert.NotEqual(t, 4, len(foundEvent.CustomQuestions), "Expected custom questions length to not be 4 (opposite assertion)")
}

// Test registering for an event with custom answers
func TestRegisterForEventWithAnswers(t *testing.T) {
	router := setupTestRouter(t)
	// First, create an event with custom questions
	eventData := gin.H{
		"title":             "Registration Test Event",
		"description":       "Test event for registration",
		"start_date":        time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"end_date":          time.Now().Add(26 * time.Hour).Format(time.RFC3339),
		"location":          "Test Venue",
		"registration_max":  10,
		"requires_approval": false,
		"custom_questions": []gin.H{
			{"question": "What is your favorite color?", "type": "text", "required": true},
			{"question": "How many years of experience do you have?", "type": "number", "required": false},
			{"question": "Select your preferred time slot", "type": "select", "required": true},
		},
	}
	w := performRequest(router, "POST", "/api/event", eventData)
	assert.Equal(t, http.StatusCreated, w.Code, "Should create event before registering")
	var createdEvent models.Event
	err := json.Unmarshal(w.Body.Bytes(), &createdEvent)
	assert.NoError(t, err)
	// Ensure the event has the questions we just created
	assert.Len(t, createdEvent.CustomQuestions, 3, "Event should have 3 custom questions to register against")

	// Now register using the real question IDs from the created event
	regBody := gin.H{
		"event_id": createdEvent.ID,
		"answers": []gin.H{
			{"question_id": createdEvent.CustomQuestions[0].ID, "answer": "Blue"},
			{"question_id": createdEvent.CustomQuestions[1].ID, "answer": "42"},
		},
	}
	w = performRequest(router, "POST", fmt.Sprintf("/api/registrations/register/%d", createdEvent.ID), regBody)
	assert.Equal(t, http.StatusCreated, w.Code, "Should successfully register with answers")
}
