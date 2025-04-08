package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SpeakerHandler struct {
	db *gorm.DB
}

func NewSpeakerHandler(db *gorm.DB) *SpeakerHandler {
	return &SpeakerHandler{db: db}
}

func (h *SpeakerHandler) Register(r *gin.RouterGroup) {
	speaker := r.Group("/speaker")
	{
		speaker.GET("", h.List)
		speaker.GET("/:id", h.Get)

		// Admin-only endpoints
		speaker.Use(middleware.AdminRequired())
		speaker.POST("", h.Create)
		speaker.PUT("/:id", h.Update)
		speaker.DELETE("/:id", h.Delete)

	}
}

func (h *SpeakerHandler) List(c *gin.Context) {
	var speaker []models.Speaker
	if err := h.db.Find(&speaker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func (h *SpeakerHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var speaker models.Speaker
	if err := h.db.First(&speaker, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func (h *SpeakerHandler) Create(c *gin.Context) {
	var speaker models.Speaker
	if err := c.ShouldBindJSON(&speaker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&speaker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func (h *SpeakerHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var speaker models.Speaker
	if err := h.db.First(&speaker, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&speaker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&speaker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, speaker)
}

func (h *SpeakerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var speaker models.Speaker
	if err := h.db.Delete(&speaker, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{"message": "Speaker deleted"})
}
