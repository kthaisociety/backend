package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SponsorHandler struct {
	db *gorm.DB
}

func NewSponsorHandler(db *gorm.DB) *SponsorHandler {
	return &SponsorHandler{db: db}
}

func (h *SponsorHandler) Register(r *gin.RouterGroup) {
	sponsor := r.Group("/sponsor")
	{
		// Public endpoints (require auth)
		sponsor.GET("", h.List)
		sponsor.GET("/:id", h.Get)

		// Admin-only endpoints
		admin := sponsor.Group("/admin")
		admin.Use(middleware.AdminRequired()) // You'll need to create this middleware
		admin.POST("", h.Create)
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)

	}
}

func (h *SponsorHandler) List(c *gin.Context) {
	var sponsor []models.Sponsor
	if err := h.db.Find(&sponsor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sponsor)
}

func (h *SponsorHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var sponsor models.Sponsor
	if err := h.db.First(&sponsor, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sponsor not found"})
		return
	}
	c.JSON(http.StatusOK, sponsor)
}

func (h *SponsorHandler) Create(c *gin.Context) {
	var sponsor models.Sponsor
	if err := c.ShouldBindJSON(&sponsor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&sponsor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sponsor)
}

func (h *SponsorHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var sponsor models.Sponsor
	if err := h.db.First(&sponsor, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sponsor not found"})
		return
	}
	if err := c.ShouldBindJSON(&sponsor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&sponsor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sponsor)
}

func (h *SponsorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var sponsor models.Sponsor
	if err := h.db.Delete(&sponsor, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sponsor deleted"})
}
