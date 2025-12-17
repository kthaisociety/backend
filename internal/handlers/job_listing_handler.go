package handlers

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobListingHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type SmallJobListing struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Company     string    `json:"company"`
	CompanyLogo string    `json:"companyLogo"`
	Salary      string    `json:"salary"`
	JobType     string    `json:"jobType"`
	Location    string    `json:"location"`
	StartDate   time.Time `json:"startdate"`
	EndDate     time.Time `json:"enddate"`
}

func NewJobListingHandler(db *gorm.DB, cfg *config.Config) *JobListingHandler {
	return &JobListingHandler{db: db, cfg: cfg}
}

func (h *JobListingHandler) Register(r *gin.RouterGroup) {
	jl := r.Group("/joblistings")
	
	// Public routes - no auth required
	jl.GET("/all", h.GetAllListings)
	jl.GET("/job", h.GetJobListing)
	
	// Admin routes - require admin role
	admin := jl.Group("/admin")
	admin.Use(middleware.RoleRequired(h.cfg, "admin"))
	{
		admin.POST("/new", h.UploadJobListing)
		admin.PUT("/update", h.UpdateJobListing)
		admin.DELETE("/delete", h.DeleteJobListing)
		admin.POST("/full", h.SingleUpload)
	}
}

// Let's make this a post
func (h *JobListingHandler) UploadJobListing(c *gin.Context) {
	var job models.JobListing
	if err := c.BindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "success", "id": job.Id})
}

// Let's make this a put
func (h *JobListingHandler) UpdateJobListing(c *gin.Context) {
	// get query params
	jobid := c.Query("id") // do we want to use jobid? or just id? We can use id for everything and maybe that is easier to remember? Not sure
	if jobid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No id provided"})
		return
	}

	jobID, err := uuid.Parse(jobid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id format"})
		return
	}

	var jl models.JobListing
	result := h.db.First(&jl, "id = ?", jobID)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job Listing not found"})
		return
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	// we know we have it, parse for updated fields
	var upjl models.JobListing
	// var upjl map[string]interface{}
	if err := c.BindJSON(&upjl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	result = h.db.Model(&jl).Updates(upjl)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// Get
func (h *JobListingHandler) GetJobListing(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No id provided"})
		return
	}

	var jl models.JobListing
	if err := h.db.First(&jl, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job listing not found"})
		return
	}
	
	// Fetch company details
	var company models.Company
	if err := h.db.First(&company, "id = ?", jl.CompanyId).Error; err != nil {
		// If company not found, still return the job but without company details
		c.JSON(http.StatusOK, jl)
		return
	}
	
	// Create a response that includes company details
	response := map[string]interface{}{
		"id":          jl.Id,
		"title":       jl.Name,
		"description": jl.Description,
		"salary":      jl.Salary,
		"location":    jl.Location,
		"jobType":     jl.JobType,
		"company":     company.Name,
		"companyLogo": company.Logo,
		"startdate":   jl.StartDate,
		"enddate":     jl.EndDate,
		"appurl":      jl.AppUrl,
		"contact":     jl.ContactInfo,
	}
	
	c.JSON(http.StatusOK, response)
}

// Get with Query Params
func (h *JobListingHandler) GetAllListings(c *gin.Context) {
	var shortListings []SmallJobListing
	
	// Build query
	query := h.db.Table("job_listings").Select(
		"job_listings.name",
		"job_listings.salary",
		"job_listings.id",
		"job_listings.job_type",
		"job_listings.location",
		"job_listings.start_date",
		"job_listings.end_date",
		"companies.name as company",
		"companies.logo as company_logo").Joins("left join companies on companies.id = job_listings.company_id")
	
	// Filter by job type if provided
	if jobType := c.Query("jobType"); jobType != "" {
		query = query.Where("job_listings.job_type = ?", jobType)
	}
	
	query.Scan(&shortListings)
	
	// Ensure we return an empty array instead of null when there are no jobs
	if shortListings == nil {
		shortListings = []SmallJobListing{}
	}
	
	c.JSON(http.StatusOK, shortListings)
}

func (h *JobListingHandler) DeleteJobListing(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No id provided"})
		return
	}
	result := h.db.Unscoped().Where("id = ?", id).Delete(&models.JobListing{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// make it easy
func (h *JobListingHandler) SingleUpload(c *gin.Context) {
	// Read JSON part into a generic map so we can accept flexible input (partial fields,
	// company name or UUID, etc.)
	// Read image (optional or required depending on endpoint contract)
	file, _ := c.FormFile("logo")
	jobFile, err := c.FormFile("job")
	if err != nil {
		c.JSON(400, gin.H{"error": "job json missing"})
		return
	}

	f, err := jobFile.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	var payload map[string]interface{}
	if err := json.NewDecoder(f).Decode(&payload); err != nil {
		c.JSON(400, gin.H{"error": "invalid job json", "detail": err.Error()})
		return
	}

	// Build JobListing from payload, handling types and company resolution
	var jl models.JobListing
	if v, ok := payload["id"].(string); ok && v != "" {
		if parsed, err := uuid.Parse(v); err == nil {
			jl.Id = parsed
		} else {
			c.JSON(400, gin.H{"error": "invalid id format"})
			return
		}
	}

	if v, ok := payload["title"].(string); ok {
		jl.Name = v
	}
	if v, ok := payload["description"].(string); ok {
		jl.Description = v
	}
	if v, ok := payload["salary"].(string); ok {
		jl.Salary = v
	}
	if v, ok := payload["location"].(string); ok {
		jl.Location = v
	}
	if v, ok := payload["jobType"].(string); ok {
		jl.JobType = v
	}
	if v, ok := payload["appurl"].(string); ok {
		jl.AppUrl = v
	}
	if v, ok := payload["startdate"].(float64); ok {
		jl.StartDate = time.Unix(int64(v), 0)
	} else {
		log.Printf("Could not parse startdate %v\n", payload["startdate"])
	}
	if v, ok := payload["enddate"].(float64); ok {
		jl.EndDate = time.Unix(int64(v), 0)
	} else {
		log.Printf("Could not parse enddate %v\n", payload["enddate"])
	}
	if v, ok := payload["contact"].(string); ok {
		jl.ContactInfo = v
	}
	var cdesc string
	if v, ok := payload["company_description"].(string); ok {
		cdesc = v
	}

	// company may exist in database
	if v, exists := payload["company"]; exists {
		cv := v.(string)
		comp, err := models.NewCompany(cv, cdesc, file, h.db, h.cfg)
		if err != nil {
			log.Printf("Error creating company: %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		jl.CompanyId = comp.Id
		if err = h.db.Create(&jl).Error; err != nil {
			log.Printf("Failed to create joblisting: %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"success": "ok"})
	}
}
