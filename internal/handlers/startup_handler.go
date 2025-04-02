package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StartupHandler struct {
	db *gorm.DB
}

func NewStartupHandler(db *gorm.DB) *StartupHandler {
	return &StartupHandler{db: db}
}

func (h *StartupHandler) Register(r *gin.RouterGroup) {
	startup := r.Group("/startup")
	{
		// Public endpoints (require auth)
		startup.GET("", h.List)
		startup.GET("/:id", h.Get)

		// Admin-only endpoints
		admin := startup.Group("/admin")
		admin.Use(middleware.AdminRequired()) // You'll need to create this middleware
		admin.POST("", h.Create)
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)

	}
}

func (h *StartupHandler) List(c *gin.Context) {
	var startup []models.Startup
	if err := h.db.Find(&startup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, startup)
}

func (h *StartupHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var startup models.Startup
	if err := h.db.First(&startup, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "startup not found"})
		return
	}
	c.JSON(http.StatusOK, startup)
}

func (h *StartupHandler) Create(c *gin.Context) {
	var startup models.Startup
	if err := c.ShouldBindJSON(&startup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&startup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, startup)
}

func (h *StartupHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var startup models.Startup
	if err := h.db.First(&startup, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Startup not found"})
		return
	}
	if err := c.ShouldBindJSON(&startup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&startup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, startup)
}

func (h *StartupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var startup models.Startup
	if err := h.db.Delete(&startup, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Startup deleted"})
}
