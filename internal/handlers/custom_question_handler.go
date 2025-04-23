package handlers

import (
	"backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomQuestionHandler struct {
	db *gorm.DB
}

func NewCustomQuestionHandler(db *gorm.DB) *CustomQuestionHandler {
	return &CustomQuestionHandler{db: db}
}

// CreateQuestion creates a new custom question for an event
func (h *CustomQuestionHandler) CreateQuestion(c *gin.Context) {
	var question models.CustomQuestion
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify event exists and belongs to the user
	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := h.db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Set the event ID
	question.EventID = uint(eventID)

	if err := h.db.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}

	c.JSON(http.StatusCreated, question)
}

// GetQuestions returns all custom questions for an event
func (h *CustomQuestionHandler) GetQuestions(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var questions []models.CustomQuestion
	if err := h.db.Where("event_id = ?", eventID).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
		return
	}

	c.JSON(http.StatusOK, questions)
}

// UpdateQuestion updates an existing custom question
func (h *CustomQuestionHandler) UpdateQuestion(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("question_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	var question models.CustomQuestion
	if err := h.db.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return
	}

	c.JSON(http.StatusOK, question)
}

// DeleteQuestion deletes a custom question
func (h *CustomQuestionHandler) DeleteQuestion(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("question_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	if err := h.db.Delete(&models.CustomQuestion{}, questionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question"})
		return
	}

	c.Status(http.StatusNoContent)
}
