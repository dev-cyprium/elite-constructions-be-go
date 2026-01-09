package handlers

import (
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/gin-gonic/gin"
)

// GetTestimonials returns paginated testimonials (10 per page)
func GetTestimonials(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// TODO: Implement with sqlc queries
	_ = db.Pool
	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    []models.Testimonial{},
		Page:    page,
		PerPage: 10,
		Total:   0,
	})
}

// GetTestimonial returns a single testimonial
func GetTestimonial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid testimonial ID")
		return
	}

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "Testimonial not found")
}

// CreateTestimonial creates a new testimonial
func CreateTestimonial(c *gin.Context) {
	var testimonial models.Testimonial
	if err := c.ShouldBindJSON(&testimonial); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Status is required
	if testimonial.Status == "" {
		ErrorResponse(c, http.StatusBadRequest, "Status is required")
		return
	}

	// TODO: Implement with sqlc queries
	_ = db.Pool
	SuccessResponse(c, http.StatusCreated, testimonial)
}

// UpdateTestimonial updates an existing testimonial
func UpdateTestimonial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid testimonial ID")
		return
	}

	var testimonial models.Testimonial
	if err := c.ShouldBindJSON(&testimonial); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "Testimonial not found")
}

// DeleteTestimonial deletes a testimonial (returns 400 if only 1 remains)
func DeleteTestimonial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid testimonial ID")
		return
	}

	// TODO: Check count - if only 1 remains, return 400
	// Otherwise delete and return 204
	_ = id
	_ = db.Pool
	
	// Placeholder
	c.Status(http.StatusNoContent)
}
