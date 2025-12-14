package handlers

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"log"
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
		admin.POST("/full", h.SingleUpload)
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

// type JobListSingle struct {
// }

// make it easy
func (h *JobListingHandler) SingleUpload(c *gin.Context) {
	// Read JSON part into a generic map so we can accept flexible input (partial fields,
	// company name or UUID, etc.)
	// Read image (optional or required depending on endpoint contract)
	file, err := c.FormFile("logo")
	has_logo := true
	if err != nil {
		// logo missing — return error if it's required
		c.JSON(400, gin.H{"error": "image required"})
		has_logo = false
		return
	}
	// read file here
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
	// if v, ok := payload["company"].(string); ok {
	// 	jl.CompanyId = v
	// }

	// company may exist in database
	if v, exists := payload["company"]; exists {
		// treat as company name; lookup in DB
		cv := v.(string)
		var comp models.Company
		if err := h.db.Where("name = ?", cv).First(&comp).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			c_id, _ := uuid.NewUUID()
			var logo_id uuid.UUID
			if has_logo {
				// create logo blob here
				r2, err := utils.InitS3SDK(h.cfg)
				if err != nil {
					log.Printf("Failed to init r2: %s\n", err)
				}
				logoBlob, err := models.NewBlobData(
					file.Filename,
					"na",
					c_id,
					fdata,
					h.db,
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
			if err = h.db.Create(&comp).Error; err != nil {
				log.Printf("Failed to create company: %s\n", err)
			}

		} else {
			log.Printf("Found Company %v\n", comp)
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
