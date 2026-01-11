package handlers

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetProjects returns paginated projects (10 per page)
func GetProjects(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()
	perPage := 10
	offset := (page - 1) * perPage

	// Get projects
	projects, err := queries.ListProjects(ctx, sqlc.ListProjectsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Get total count
	total, err := queries.CountProjects(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map sqlc projects to models and load images
	projectModels := make([]models.Project, len(projects))
	for i, p := range projects {
		projectModels[i] = mapSQLCProjectToModel(p)
		// Load images for each project
		images, err := queries.ListProjectImagesByProjectID(ctx, p.ID)
		if err == nil {
			projectModels[i].Images = mapSQLCProjectImagesToModels(images)
		}
	}

	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    projectModels,
		Page:    page,
		PerPage: perPage,
		Total:   total,
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Get project
	project, err := queries.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Get project images
	images, err := queries.ListProjectImagesByProjectID(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	projectModel := mapSQLCProjectToModel(project)
	projectModel.Images = mapSQLCProjectImagesToModels(images)

	SuccessResponse(c, http.StatusOK, projectModel)
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

		queries := sqlc.New(db.Pool)
		ctx := c.Request.Context()

		// Start transaction
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to start transaction")
			return
		}
		defer tx.Rollback(ctx)

		qtx := queries.WithTx(tx)

		// Create project in database
		highlighted := highlightImageIndex >= 0 && highlightImageIndex < len(files)
		project, err := qtx.CreateProject(ctx, sqlc.CreateProjectParams{
			Status:      1,
			Name:        name,
			Category:    pgtypeTextPtr(category),
			Client:      pgtypeTextPtr(client),
			Order:       int32(order),
			Highlighted: highlighted,
		})
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to create project")
			return
		}

		// Save files, generate blurhash, create project_images records
		for i, file := range files {
			// Open file
			src, err := file.Open()
			if err != nil {
				ErrorResponse(c, http.StatusBadRequest, "Failed to open file", err.Error())
				return
			}

			// Read file data
			fileData, err := io.ReadAll(src)
			src.Close()
			if err != nil {
				ErrorResponse(c, http.StatusBadRequest, "Failed to read file", err.Error())
				return
			}

			// Save file
			url, err := storage.SaveFile(fileData, file.Filename, cfg.StoragePath)
			if err != nil {
				ErrorResponse(c, http.StatusBadRequest, "Failed to save file", err.Error())
				return
			}

			// Generate blurhash
			filePath := filepath.Join(cfg.StoragePath, "public", "img", filepath.Base(url))
			blurHash, err := storage.GenerateBlurHash(filePath)
			var blurHashPtr pgtype.Text
			if err == nil {
				blurHashPtr = pgtype.Text{String: blurHash, Valid: true}
			}

			// Create project image
			_, err = qtx.CreateProjectImage(ctx, sqlc.CreateProjectImageParams{
				Name:      file.Filename,
				Url:       url,
				ProjectID: project.ID,
				Order:     int32(i),
				BlurHash:  blurHashPtr,
			})
			if err != nil {
				ErrorResponse(c, http.StatusInternalServerError, "Failed to create project image")
				return
			}
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		// Get project with images
		images, _ := queries.ListProjectImagesByProjectID(ctx, project.ID)
		projectModel := mapSQLCProjectToModel(project)
		projectModel.Images = mapSQLCProjectImagesToModels(images)

		SuccessResponse(c, http.StatusCreated, projectModel)
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
		existingImageIDsStr := form.Value["files[]"] // Can contain IDs or be empty
		newFiles := form.File["files[]"]

		queries := sqlc.New(db.Pool)
		ctx := c.Request.Context()

		// Check if project exists
		project, err := queries.GetProjectByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ErrorResponse(c, http.StatusNotFound, "Project not found")
				return
			}
			ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}

		// Get current image IDs
		currentImages, err := queries.ListProjectImageIDsByProjectID(ctx, id)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}

		// Parse existing image IDs from form
		existingImageIDs := make(map[int64]bool)
		for _, idStr := range existingImageIDsStr {
			if imgID, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				existingImageIDs[imgID] = true
			}
		}

		// Start transaction
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to start transaction")
			return
		}
		defer tx.Rollback(ctx)

		qtx := queries.WithTx(tx)

		// Determine highlighted status
		highlighted := project.Highlighted
		if highlightImageIndex >= 0 {
			// If highlightImageIndex is provided, it refers to the final image order
			totalImages := len(existingImageIDs) + len(newFiles)
			highlighted = highlightImageIndex < totalImages
		}

		// Update project
		err = qtx.UpdateProject(ctx, sqlc.UpdateProjectParams{
			ID:          id,
			Status:      int16(project.Status),
			Name:        name,
			Category:    pgtypeTextPtr(category),
			Client:      pgtypeTextPtr(client),
			Order:       int32(order),
			Highlighted: highlighted,
		})
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
			return
		}

		// Delete removed images
		for _, imgID := range currentImages {
			if !existingImageIDs[imgID] {
				// Get image to get URL for file deletion
				img, err := qtx.GetProjectImageByID(ctx, imgID)
				if err == nil {
					// Delete from database
					err = qtx.DeleteProjectImage(ctx, imgID)
					if err == nil {
						// Delete file from filesystem
						storage.DeleteFile(img.Url, cfg.StoragePath)
					}
				}
			}
		}

		// Add new files
		nextOrder := len(existingImageIDs)
		for _, file := range newFiles {
			// Open file
			src, err := file.Open()
			if err != nil {
				ErrorResponse(c, http.StatusBadRequest, "Failed to open file", err.Error())
				return
			}

			// Read file data
			fileData, err := io.ReadAll(src)
			src.Close()
			if err != nil {
				ErrorResponse(c, http.StatusBadRequest, "Failed to read file", err.Error())
				return
			}

			// Save file
			url, err := storage.SaveFile(fileData, file.Filename, cfg.StoragePath)
			if err != nil {
				ErrorResponse(c, http.StatusBadRequest, "Failed to save file", err.Error())
				return
			}

			// Generate blurhash
			filePath := filepath.Join(cfg.StoragePath, "public", "img", filepath.Base(url))
			blurHash, err := storage.GenerateBlurHash(filePath)
			var blurHashPtr pgtype.Text
			if err == nil {
				blurHashPtr = pgtype.Text{String: blurHash, Valid: true}
			}

			// Create project image
			_, err = qtx.CreateProjectImage(ctx, sqlc.CreateProjectImageParams{
				Name:      file.Filename,
				Url:       url,
				ProjectID: id,
				Order:     int32(nextOrder),
				BlurHash:  blurHashPtr,
			})
			if err != nil {
				ErrorResponse(c, http.StatusInternalServerError, "Failed to create project image")
				return
			}
			nextOrder++
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		// Get updated project with images
		updatedProject, _ := queries.GetProjectByID(ctx, id)
		images, _ := queries.ListProjectImagesByProjectID(ctx, id)
		projectModel := mapSQLCProjectToModel(updatedProject)
		projectModel.Images = mapSQLCProjectImagesToModels(images)

		SuccessResponse(c, http.StatusOK, projectModel)
	}
}

// GetProjectImage returns a single project image by ID
func GetProjectImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid project image ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Get project image
	image, err := queries.GetProjectImageByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Project image not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map to model
	var blurHash *string
	if image.BlurHash.Valid {
		blurHash = &image.BlurHash.String
	}

	imageModel := models.ProjectImage{
		ID:        image.ID,
		Name:      image.Name,
		URL:       image.Url,
		ProjectID: image.ProjectID,
		Order:     int(image.Order),
		BlurHash:  blurHash,
		CreatedAt: image.CreatedAt.Time,
		UpdatedAt: image.UpdatedAt.Time,
	}

	SuccessResponse(c, http.StatusOK, imageModel)
}

// ToggleHighlight toggles the highlighted boolean
func ToggleHighlight(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if project exists
	_, err = queries.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Toggle highlight
	err = queries.ToggleProjectHighlight(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to toggle highlight")
		return
	}

	// Get updated project
	project, err := queries.GetProjectByID(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	SuccessResponse(c, http.StatusOK, mapSQLCProjectToModel(project))
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

		queries := sqlc.New(db.Pool)
		ctx := c.Request.Context()

		// Check if project exists
		_, err = queries.GetProjectByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ErrorResponse(c, http.StatusNotFound, "Project not found")
				return
			}
			ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}

		// Get project images before deletion (to delete files)
		images, err := queries.DeleteProjectImagesByProjectID(ctx, id)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}

		// Delete image files from filesystem
		for _, img := range images {
			storage.DeleteFile(img.Url, cfg.StoragePath)
		}

		// Delete project (cascade will delete images from DB, but we already got them)
		err = queries.DeleteProject(ctx, id)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to delete project")
			return
		}

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

func pgtypeTextPtr(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// Helper function to map sqlc Project to models.Project
func mapSQLCProjectToModel(p sqlc.Project) models.Project {
	var category *string
	if p.Category.Valid {
		category = &p.Category.String
	}

	var client *string
	if p.Client.Valid {
		client = &p.Client.String
	}

	return models.Project{
		ID:          p.ID,
		Status:      int(p.Status),
		Name:        p.Name,
		Category:    category,
		Client:      client,
		Order:       int(p.Order),
		Highlighted: p.Highlighted,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}

// Helper function to map sqlc ProjectImages to models.ProjectImage
func mapSQLCProjectImagesToModels(images []sqlc.ProjectImage) []models.ProjectImage {
	result := make([]models.ProjectImage, len(images))
	for i, img := range images {
		var blurHash *string
		if img.BlurHash.Valid {
			blurHash = &img.BlurHash.String
		}

		result[i] = models.ProjectImage{
			ID:        img.ID,
			Name:      img.Name,
			URL:       img.Url,
			ProjectID: img.ProjectID,
			Order:     int(img.Order),
			BlurHash:  blurHash,
			CreatedAt: img.CreatedAt.Time,
			UpdatedAt: img.UpdatedAt.Time,
		}
	}
	return result
}
