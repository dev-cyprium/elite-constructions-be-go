package handlers

import (
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/gin-gonic/gin"
)

// GetVisitorMessages returns paginated visitor messages (10 per page)
func GetVisitorMessages(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()
	perPage := 10
	offset := (page - 1) * perPage

	messages, err := queries.ListVisitorMessages(ctx, sqlc.ListVisitorMessagesParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	total, err := queries.CountVisitorMessages(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	messageModels := make([]models.VisitorMessage, len(messages))
	for i, m := range messages {
		messageModels[i] = mapSQLCVisitorMessageToModel(m)
	}

	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    messageModels,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// DeleteVisitorMessage deletes a visitor message (returns 204)
func DeleteVisitorMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid visitor message ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if message exists (optional, but good for error handling)
	// We don't have a GetVisitorMessageByID query, so we'll just try to delete
	err = queries.DeleteVisitorMessage(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete visitor message")
		return
	}

	c.Status(http.StatusNoContent)
}
