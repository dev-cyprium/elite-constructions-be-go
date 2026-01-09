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

// GetTestimonials returns paginated testimonials (10 per page)
func GetTestimonials(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()
	perPage := 10
	offset := (page - 1) * perPage

	testimonials, err := queries.ListTestimonials(ctx, sqlc.ListTestimonialsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	total, err := queries.CountTestimonials(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	testimonialModels := make([]models.Testimonial, len(testimonials))
	for i, t := range testimonials {
		testimonialModels[i] = mapSQLCTestimonialToModel(t)
	}

	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    testimonialModels,
		Page:    page,
		PerPage: perPage,
		Total:   total,
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	testimonial, err := queries.GetTestimonialByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Testimonial not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	SuccessResponse(c, http.StatusOK, mapSQLCTestimonialToModel(testimonial))
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if testimonial exists
	_, err = queries.GetTestimonialByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Testimonial not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Update testimonial
	err = queries.UpdateTestimonial(ctx, sqlc.UpdateTestimonialParams{
		ID:          id,
		FullName:    testimonial.FullName,
		Profession:  testimonial.Profession,
		Testimonial: testimonial.Testimonial,
		Status:      testimonial.Status,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update testimonial")
		return
	}

	// Get updated testimonial
	updated, err := queries.GetTestimonialByID(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	SuccessResponse(c, http.StatusOK, mapSQLCTestimonialToModel(updated))
}

// DeleteTestimonial deletes a testimonial (returns 400 if only 1 remains)
func DeleteTestimonial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid testimonial ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check count - if only 1 remains, return 400
	count, err := queries.CountTestimonials(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	if count <= 1 {
		ErrorResponse(c, http.StatusBadRequest, "Cannot delete the last testimonial")
		return
	}

	// Check if testimonial exists
	_, err = queries.GetTestimonialByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Testimonial not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Delete testimonial
	err = queries.DeleteTestimonial(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete testimonial")
		return
	}

	c.Status(http.StatusNoContent)
}
