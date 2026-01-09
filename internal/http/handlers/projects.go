package handlers

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/gin-gonic/gin"
)

// GetProjects returns paginated projects (10 per page)
func GetProjects(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// TODO: Implement with sqlc queries
	_ = db.Pool
	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    []models.Project{},
		Page:    page,
		PerPage: 10,
		Total:   0,
	})
}

// GetProject returns a single project with images
func GetProject(c *gin.Context) {
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

// CreateProject creates a new project with multipart file uploads
func CreateProject(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid multipart form", err.Error())
			return
		}

		// Extract form fields
		name := getFormValue(form, "name")
		if name == "" {
			ErrorResponse(c, http.StatusBadRequest, "Name is required")
			return
		}

		category := getFormValue(form, "category")
		client := getFormValue(form, "client")
		orderStr := getFormValue(form, "order")
		order := 0
		if orderStr != "" {
			order, _ = strconv.Atoi(orderStr)
		}

		highlightImageIndexStr := getFormValue(form, "highlightImageIndex")
		highlightImageIndex := -1
		if highlightImageIndexStr != "" {
			highlightImageIndex, _ = strconv.Atoi(highlightImageIndexStr)
		}

		// Get uploaded files
		files := form.File["files[]"]
		if len(files) == 0 {
			ErrorResponse(c, http.StatusBadRequest, "At least one file is required")
			return
		}

		// TODO: Create project in database
		// TODO: Save files, generate blurhash, create project_images records
		_ = db.Pool
		_ = cfg

		project := models.Project{
			Name:        name,
			Category:    stringPtr(category),
			Client:      stringPtr(client),
			Order:       order,
			Highlighted: highlightImageIndex >= 0 && highlightImageIndex < len(files),
		}

		SuccessResponse(c, http.StatusCreated, project)
	}
}

// UpdateProject updates a project with multipart file uploads
func UpdateProject(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
			return
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid multipart form", err.Error())
			return
		}

		// Extract form fields
		name := getFormValue(form, "name")
		category := getFormValue(form, "category")
		client := getFormValue(form, "client")
		orderStr := getFormValue(form, "order")
		order := 0
		if orderStr != "" {
			order, _ = strconv.Atoi(orderStr)
		}

		highlightImageIndexStr := getFormValue(form, "highlightImageIndex")
		highlightImageIndex := -1
		if highlightImageIndexStr != "" {
			highlightImageIndex, _ = strconv.Atoi(highlightImageIndexStr)
		}

		// Get existing image IDs and new files
		existingImageIDs := form.Value["files[]"] // Can contain IDs or be empty
		newFiles := form.File["files[]"]

		// TODO: Update project in database
		// TODO: Handle existing image IDs (keep them)
		// TODO: Save new files, generate blurhash, create project_images records
		// TODO: Delete removed images (compare existing with provided IDs)
		_ = id
		_ = db.Pool
		_ = cfg
		_ = existingImageIDs
		_ = newFiles
		_ = name
		_ = category
		_ = client
		_ = order
		_ = highlightImageIndex

		ErrorResponse(c, http.StatusNotFound, "Project not found")
	}
}

// ToggleHighlight toggles the highlighted boolean
func ToggleHighlight(c *gin.Context) {
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

// DeleteProject deletes a project (cascade deletes images via FK)
func DeleteProject(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
			return
		}

		// TODO: Get project images before deletion
		// TODO: Delete project (cascade will delete images)
		// TODO: Delete image files from filesystem
		_ = id
		_ = db.Pool
		_ = cfg

		c.Status(http.StatusNoContent)
	}
}

// Helper functions
func getFormValue(form *multipart.Form, key string) string {
	if values, ok := form.Value[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
