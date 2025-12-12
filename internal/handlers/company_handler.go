package handlers

import (
	"backend/internal/config"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CompanyHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewCompanyHandler(db *gorm.DB, cfg *config.Config) *CompanyHandler {
	return &CompanyHandler{db: db, cfg: cfg}
}

func (h *CompanyHandler) Register(r *gin.RouterGroup) {
	companies := r.Group("/company")
	{
		// Define company-related routes here
		_ = companies.POST("/addCompany", h.UploadCompany)
		// upload.Use(middleware.RoleRequired(h.cfg, "admin"))
		_ = companies.GET("/getCompany", h.GetCompany)
		_ = companies.GET("/getAllCompanies", h.GetAllCompanies)
	}
}

func (h *CompanyHandler) UploadCompany(c *gin.Context) {
	var companyData models.Company
	if err := c.BindJSON(&companyData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&companyData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusAccepted, "Company added successfully")
}

func (h *CompanyHandler) GetCompany(c *gin.Context) {
	// Implementation for getting a single company
	id := c.Query("id")
	var company models.Company
	if id != "" {
		if err := h.db.First(&company, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		c.JSON(http.StatusOK, company)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
	}
}

func (h *CompanyHandler) GetAllCompanies(c *gin.Context) {
	// Implementation for getting all companies
	var companies []models.Company
	h.db.Table("companies").Select(
		"companies.id",
		"companies.name").Scan(&companies)
	c.JSON(http.StatusOK, companies)
}
