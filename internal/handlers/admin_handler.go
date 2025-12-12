package handlers

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/utils"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAdminHandler(db *gorm.DB, cfg *config.Config) *AdminHandler {
	return &AdminHandler{db: db, cfg: cfg}
}

func (h *AdminHandler) Register(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.RoleRequired(h.cfg, "admin"))
	{
		// Auth required endpoints
		admin.Use(middleware.AuthRequiredJWT(h.cfg))
		admin.GET("/listadmins", h.ListAdmins)
		admin.PUT("/setadmin", h.PromoteToAdmin)
	}
	checkAdmin := r.Group("/checkadmin")
	{
		checkAdmin.GET("/", h.IsAdmin)
	}
}

func (h *AdminHandler) IsAdmin(c *gin.Context) {
	retValid := func(isAd bool) {
		c.JSON(200, gin.H{"is_admin": isAd})
	}
	// Implementation for checking if the user is an admin
	for _, cookie := range c.Request.Cookies() {
		if cookie.Name == "jwt" {
			valid, token := utils.ParseAndVerify(cookie.Value, h.cfg.JwtSigningKey)
			if !valid {
				retValid(false)
			}
			roles := strings.Split(utils.GetClaims(token)["roles"].(string), ",")
			if slices.Contains(roles, "admin") {
				retValid(true)
			} else {
				retValid(false)
			}
		}
	}
}

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	// Implementation for listing all admin users
	var admins []models.User
	if err := h.db.Where("role = ?", "admin").Find(&admins).Error; err != nil {
		c.JSON(500, gin.H{"error": "Could not retrieve admins"})
		return
	}
	c.JSON(200, admins)
}

func (h *AdminHandler) PromoteToAdmin(c *gin.Context) {
	// Implementation for promoting a user to admin
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{req.UserID: "Invalid request"})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(400, gin.H{req.UserID: "Invalid user ID"})
		return
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("role", "admin").Error; err != nil {
		c.JSON(500, gin.H{req.UserID: "failure"})
		return
	}
	c.JSON(200, gin.H{req.UserID: "success"})
}
