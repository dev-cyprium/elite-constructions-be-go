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

type UpdateStaticTextRequest struct {
	Content string `json:"content" binding:"required"`
}

// GetStaticTexts returns paginated static texts (10 per page)
func GetStaticTexts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()
	perPage := 10
	offset := (page - 1) * perPage

	staticTexts, err := queries.ListStaticTexts(ctx, sqlc.ListStaticTextsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	total, err := queries.CountStaticTexts(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	staticTextModels := make([]models.StaticText, len(staticTexts))
	for i, st := range staticTexts {
		staticTextModels[i] = mapSQLCStaticTextToModel(st)
	}

	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    staticTextModels,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GetStaticText returns a single static text
func GetStaticText(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid static text ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	staticText, err := queries.GetStaticTextByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Static text not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	SuccessResponse(c, http.StatusOK, mapSQLCStaticTextToModel(staticText))
}

// UpdateStaticText updates static text content only (key/label immutable)
func UpdateStaticText(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid static text ID")
		return
	}

	var req UpdateStaticTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if static text exists
	_, err = queries.GetStaticTextByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Static text not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Update content
	err = queries.UpdateStaticTextContent(ctx, sqlc.UpdateStaticTextContentParams{
		ID:      id,
		Content: req.Content,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update static text")
		return
	}

	// Get updated static text
	updated, err := queries.GetStaticTextByID(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	SuccessResponse(c, http.StatusOK, mapSQLCStaticTextToModel(updated))
}
