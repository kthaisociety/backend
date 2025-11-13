package handlers

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobListingHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewJobListingHandler(db *gorm.DB, cfg *config.Config) *JobListingHandler {
	return &JobListingHandler{db: db, cfg: cfg}
}

func (h *JobListingHandler) Register(r *gin.RouterGroup) {
	jl := r.Group("/joblistings")
	{
		post := jl.POST("/new", h.UploadJobListing)
		put := jl.PUT("/update", h.UpdateJobListing)
		// require admin to upload
		post.Use(middleware.RoleRequired(h.cfg, "admin"))
		put.Use(middleware.RoleRequired(h.cfg, "admin"))
		// no auth required for these
		jl.GET("/all", h.GetAllListing)
		jl.GET("/", h.GetJobListings)
	}
}

// Let's make this a post
func (h *JobListingHandler) UploadJobListing(c *gin.Context) {
	var job models.JobListing
	if err := c.BindJSON(job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if err := h.db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.String(http.StatusAccepted, "success")
}

// Let's make this a put
func (h *JobListingHandler) UpdateJobListing(c *gin.Context) {
	// get query params
	jobid := c.Query("id") // do we want to use jobid? or just id? We can use id for everything and maybe that is easier to remember? Not sure
	if jobid != "" {
		var jl models.JobListing
		result := h.db.First(&jl, "id = ?", jobid)
		if result.Error == gorm.ErrRecordNotFound {
			c.String(http.StatusBadRequest, "Job Listing not found")
		} else if result.Error != nil {
			c.String(http.StatusInternalServerError, "Error accessing data %s\n", result.Error)
		}
		// we know we have it, parse for updated fields
		var upjl models.JobListing
		// var upjl map[string]interface{}
		c.BindJSON(&upjl)
		result = h.db.Model(&jl).Updates(upjl)
		if result.Error != nil {
			c.String(http.StatusInternalServerError, "Could not update Job Listing %s\n", result.Error)
		}
		c.String(http.StatusAccepted, "Success")
	}
}

// Get
func (h *JobListingHandler) GetAllListing(c *gin.Context) {
	var jls models.JobListing[]
	type result struct {
		name string
		company string
		salary string
		id uuid.UUID
	}
	// h.db.Select("id", "name", "salary").Find(&jls)
	h.db.Table("job_listings").Select(
		"job_listings.name",
		"job_listings.salary",
		"job_listings.id", 
		"company.name").Joins("left join companies on company.id = job_listing.id")
}

// Get with Query Params
func (h *JobListingHandler) GetJobListings(c *gin.Context) {

}
