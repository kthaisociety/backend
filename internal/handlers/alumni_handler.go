package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlumniHandler struct {
	db *gorm.DB
}

func NewAlumniHandler(db *gorm.DB) *AlumniHandler {
	return &AlumniHandler{db: db}
}

func (h *AlumniHandler) Register(r *gin.RouterGroup) {
	alumni := r.Group("/alumni")
	{
		alumni.GET("", h.List)
		alumni.GET("/:id", h.Get)

		// Admin-only endpoints
		alumni.Use(middleware.AdminRequired())
		alumni.POST("", h.Create)
		alumni.PUT("/:id", h.Update)
		alumni.DELETE("/:id", h.Delete)

	}
}

func (h *AlumniHandler) List(c *gin.Context) {
	var alumni []models.Alumni
	if err := h.db.Find(&alumni).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alumni)
}

func (h *AlumniHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var alumni models.Alumni
	if err := h.db.First(&alumni, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alumni)
}

func (h *AlumniHandler) Create(c *gin.Context) {
	var alumni models.Alumni
	if err := c.ShouldBindJSON(&alumni); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&alumni).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alumni)
}

func (h *AlumniHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var alumni models.Alumni
	if err := h.db.First(&alumni, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&alumni); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&alumni).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alumni)
}

func (h *AlumniHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var alumni models.Alumni
	if err := h.db.Delete(&alumni, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{"message": "Alumni deleted"})
}
