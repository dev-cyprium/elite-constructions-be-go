package handlers

import (
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/gin-gonic/gin"
)

// Ping returns a welcome message
func Ping(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Welcome to API 1.0"})
}

// GetPublicProjects returns all projects (no pagination)
func GetPublicProjects(c *gin.Context) {
	// TODO: Implement with sqlc queries
	// For now, return empty array
	SuccessResponse(c, http.StatusOK, gin.H{"data": []models.Project{}})
}

// GetHighlightedProjects returns only highlighted projects
func GetHighlightedProjects(c *gin.Context) {
	// TODO: Implement with sqlc queries
	SuccessResponse(c, http.StatusOK, gin.H{"data": []models.Project{}})
}

// GetPublicProjectsPaginated returns paginated projects (3 per page)
func GetPublicProjectsPaginated(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// TODO: Implement with sqlc queries
	// Per page: 3
	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    []models.Project{},
		Page:    page,
		PerPage: 3,
		Total:   0,
	})
}

// GetPublicProject returns a single project with images
func GetPublicProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "Project not found")
}

// GetPublicTestimonials returns only testimonials with status='ready'
func GetPublicTestimonials(c *gin.Context) {
	// TODO: Implement with sqlc queries
	SuccessResponse(c, http.StatusOK, gin.H{"data": []models.Testimonial{}})
}

// CreatePublicTestimonial creates a new testimonial with status='pending'
func CreatePublicTestimonial(c *gin.Context) {
	var testimonial models.Testimonial
	if err := c.ShouldBindJSON(&testimonial); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set status to pending
	testimonial.Status = "pending"

	// TODO: Implement with sqlc queries to insert
	SuccessResponse(c, http.StatusCreated, testimonial)
}

// GetPublicStaticTexts returns all static texts
func GetPublicStaticTexts(c *gin.Context) {
	// TODO: Implement with sqlc queries
	SuccessResponse(c, http.StatusOK, gin.H{"data": []models.StaticText{}})
}

// GetPublicConfigs returns all configurations
func GetPublicConfigs(c *gin.Context) {
	// TODO: Implement with sqlc queries
	SuccessResponse(c, http.StatusOK, gin.H{"data": []models.Configuration{}})
}

// CreateVisitorMessage creates a new visitor message
func CreateVisitorMessage(c *gin.Context) {
	var message models.VisitorMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// TODO: Implement with sqlc queries to insert
	SuccessResponse(c, http.StatusCreated, message)
}
