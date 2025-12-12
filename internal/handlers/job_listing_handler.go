package handlers

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobListingHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type SmallJobListing struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"title"`
	Company string    `json:"company"`
	Salary  string    `json:"salary"`
}

func NewJobListingHandler(db *gorm.DB, cfg *config.Config) *JobListingHandler {
	return &JobListingHandler{db: db, cfg: cfg}
}

func (h *JobListingHandler) Register(r *gin.RouterGroup) {
	jl := r.Group("/joblistings")
	admin := jl.Group("/admin")
	admin.Use(middleware.RoleRequired(h.cfg, "admin"))
	{
		admin.POST("/new", h.UploadJobListing)
		admin.PUT("/update", h.UpdateJobListing)
		admin.DELETE("/delete", h.DeleteJobListing)
		// no auth required for these
		jl.GET("/all", h.GetAllListings)
		jl.GET("/job", h.GetJobListing)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	c.JSON(http.StatusOK, jl)
}

// Get with Query Params
func (h *JobListingHandler) GetAllListings(c *gin.Context) {
	var shortListings []SmallJobListing
	h.db.Table("job_listings").Select(
		"job_listings.name",
		"job_listings.salary",
		"job_listings.id",
		"companies.name as company").Joins("left join companies on companies.id = job_listings.company_id").Scan(&shortListings)
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
