package handlers

import (
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/gin-gonic/gin"
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

	// TODO: Implement with sqlc queries
	_ = db.Pool
	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    []models.StaticText{},
		Page:    page,
		PerPage: 10,
		Total:   0,
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

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "Static text not found")
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

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "Static text not found")
}
