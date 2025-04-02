package handlers

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeamMemberHandler struct {
	db *gorm.DB
}

func NewTeamMemberHandler(db *gorm.DB) *TeamMemberHandler {
	return &TeamMemberHandler{db: db}
}

func (h *TeamMemberHandler) Register(r *gin.RouterGroup) {
	teamMembers := r.Group("/team-members")
	{
		// Public endpoints
		teamMembers.GET("", h.List)
		teamMembers.GET("/:id", h.Get)

		// Admin-only endpoints
		admin := teamMembers.Group("/admin")
		admin.Use(middleware.AdminRequired())
		admin.POST("", h.Create)
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)
	}
}

func (h *TeamMemberHandler) List(c *gin.Context) {
	var teamMembers []models.TeamMember
	if err := h.db.Find(&teamMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, teamMembers)
}

func (h *TeamMemberHandler) Create(c *gin.Context) {
	var teamMember models.TeamMember
	if err := c.ShouldBindJSON(&teamMember); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&teamMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, teamMember)
}

func (h *TeamMemberHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var teamMember models.TeamMember
	if err := h.db.First(&teamMember, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}
	c.JSON(http.StatusOK, teamMember)
}

func (h *TeamMemberHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var teamMember models.TeamMember
	if err := h.db.First(&teamMember, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	if err := c.ShouldBindJSON(&teamMember); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&teamMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, teamMember)
}

func (h *TeamMemberHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var teamMember models.TeamMember
	if err := h.db.First(&teamMember, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	if err := h.db.Delete(&teamMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
