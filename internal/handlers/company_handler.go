package handlers

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/utils"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func NewCompany(cv string, description string, file multipart.FileHeader, db *gorm.DB, cfg *config.Config) (*models.Company, error) {
	// read file here
	has_logo := true
	fdata := make([]byte, file.Size)
	f_reader, _ := file.Open()
	nread, err := f_reader.Read(fdata)
	if err != nil {
		log.Printf("Could not Read logo file: %s\n", err)
		has_logo = false
	}
	if nread != int(file.Size) {
		log.Printf("Read wrong number of bytes Read: %v -- File: %v\n", nread, file.Size)
		has_logo = false
	}
	var comp models.Company
	if err := db.Where("name = ?", cv).First(&comp).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}

		c_id, _ := uuid.NewUUID()
		var logo_id uuid.UUID
		if has_logo {
			// create logo blob here
			r2, err := utils.InitS3SDK(cfg)
			if err != nil {
				log.Printf("Failed to init r2: %s\n", err)
				return nil, err
			}
			logoBlob, err := models.NewBlobData(
				file.Filename,
				"na",
				c_id,
				fdata,
				db,
				r2,
			)
			logo_id = logoBlob.BlobId
		}
		// create new company here
		comp = models.Company{
			Id:          c_id,
			Name:        cv,
			Description: "",
			Logo:        logo_id,
		}
		if err = db.Create(&comp).Error; err != nil {
			log.Printf("Failed to create company: %s\n", err)
			return nil, err
		}

	} else {
		log.Printf("Found Company %v\n", comp)
	}
	return &comp, nil
}

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
