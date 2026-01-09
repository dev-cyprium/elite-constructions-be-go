package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// Ping returns a welcome message
func Ping(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Welcome to API 1.0"})
}

// GetPublicProjects returns all projects (no pagination)
func GetPublicProjects(c *gin.Context) {
	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	projects, err := queries.ListPublicProjects(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map and load images
	projectModels := make([]models.Project, len(projects))
	for i, p := range projects {
		projectModels[i] = mapSQLCProjectToModel(p)
		images, err := queries.ListProjectImagesByProjectID(ctx, p.ID)
		if err == nil {
			projectModels[i].Images = mapSQLCProjectImagesToModels(images)
		}
	}

	SuccessResponse(c, http.StatusOK, gin.H{"data": projectModels})
}

// GetHighlightedProjects returns only highlighted projects
func GetHighlightedProjects(c *gin.Context) {
	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	projects, err := queries.ListHighlightedProjects(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map and load images
	projectModels := make([]models.Project, len(projects))
	for i, p := range projects {
		projectModels[i] = mapSQLCProjectToModel(p)
		images, err := queries.ListProjectImagesByProjectID(ctx, p.ID)
		if err == nil {
			projectModels[i].Images = mapSQLCProjectImagesToModels(images)
		}
	}

	SuccessResponse(c, http.StatusOK, gin.H{"data": projectModels})
}

// GetPublicProjectsPaginated returns paginated projects (3 per page)
func GetPublicProjectsPaginated(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()
	perPage := 3
	offset := (page - 1) * perPage

	projects, err := queries.ListPublicProjectsPaginated(ctx, sqlc.ListPublicProjectsPaginatedParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	total, err := queries.CountProjects(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map and load images
	projectModels := make([]models.Project, len(projects))
	for i, p := range projects {
		projectModels[i] = mapSQLCProjectToModel(p)
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

// GetPublicProject returns a single project with images
func GetPublicProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	project, err := queries.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	images, err := queries.ListProjectImagesByProjectID(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	projectModel := mapSQLCProjectToModel(project)
	projectModel.Images = mapSQLCProjectImagesToModels(images)

	SuccessResponse(c, http.StatusOK, projectModel)
}

// GetPublicTestimonials returns only testimonials with status='ready'
func GetPublicTestimonials(c *gin.Context) {
	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	testimonials, err := queries.ListPublicTestimonials(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	testimonialModels := make([]models.Testimonial, len(testimonials))
	for i, t := range testimonials {
		testimonialModels[i] = mapSQLCTestimonialToModel(t)
	}

	SuccessResponse(c, http.StatusOK, gin.H{"data": testimonialModels})
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	created, err := queries.CreateTestimonial(ctx, sqlc.CreateTestimonialParams{
		FullName:    testimonial.FullName,
		Profession:  testimonial.Profession,
		Testimonial: testimonial.Testimonial,
		Status:      testimonial.Status,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create testimonial")
		return
	}

	SuccessResponse(c, http.StatusCreated, mapSQLCTestimonialToModel(created))
}

// GetPublicStaticTexts returns all static texts
func GetPublicStaticTexts(c *gin.Context) {
	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	staticTexts, err := queries.ListAllStaticTexts(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	staticTextModels := make([]models.StaticText, len(staticTexts))
	for i, st := range staticTexts {
		staticTextModels[i] = mapSQLCStaticTextToModel(st)
	}

	SuccessResponse(c, http.StatusOK, gin.H{"data": staticTextModels})
}

// GetPublicConfigs returns all configurations
func GetPublicConfigs(c *gin.Context) {
	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	configs, err := queries.ListAllConfigurations(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	configModels := make([]models.Configuration, len(configs))
	for i, cfg := range configs {
		configModels[i] = mapSQLCConfigurationToModel(cfg)
	}

	SuccessResponse(c, http.StatusOK, gin.H{"data": configModels})
}

// CreateVisitorMessage creates a new visitor message
func CreateVisitorMessage(c *gin.Context) {
	var message models.VisitorMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	created, err := queries.CreateVisitorMessage(ctx, sqlc.CreateVisitorMessageParams{
		Email:       message.Email,
		Address:     message.Address,
		Description: message.Description,
		Seen:        message.Seen,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create visitor message")
		return
	}

	SuccessResponse(c, http.StatusCreated, mapSQLCVisitorMessageToModel(created))
}

// Helper functions for mapping sqlc types to models
func mapSQLCTestimonialToModel(t sqlc.Testimonial) models.Testimonial {
	return models.Testimonial{
		ID:          t.ID,
		FullName:    t.FullName,
		Profession:  t.Profession,
		Testimonial: t.Testimonial,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Time,
		UpdatedAt:   t.UpdatedAt.Time,
	}
}

func mapSQLCStaticTextToModel(st sqlc.StaticText) models.StaticText {
	return models.StaticText{
		ID:        st.ID,
		Key:       st.Key,
		Label:     st.Label,
		Content:   st.Content,
		CreatedAt: st.CreatedAt.Time,
		UpdatedAt: st.UpdatedAt.Time,
	}
}

func mapSQLCConfigurationToModel(cfg sqlc.Configuration) models.Configuration {
	return models.Configuration{
		ID:        cfg.ID,
		Key:       cfg.Key,
		Value:     cfg.Value,
		CreatedAt: cfg.CreatedAt.Time,
		UpdatedAt: cfg.UpdatedAt.Time,
	}
}

func mapSQLCVisitorMessageToModel(vm sqlc.VisitorMessage) models.VisitorMessage {
	return models.VisitorMessage{
		ID:          vm.ID,
		Email:       vm.Email,
		Address:     vm.Address,
		Description: vm.Description,
		Seen:        vm.Seen,
		CreatedAt:   vm.CreatedAt.Time,
		UpdatedAt:   vm.UpdatedAt.Time,
	}
}
