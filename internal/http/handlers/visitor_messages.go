package handlers

import (
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/gin-gonic/gin"
)

// GetVisitorMessages returns paginated visitor messages (10 per page)
func GetVisitorMessages(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// TODO: Implement with sqlc queries
	_ = db.Pool
	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    []models.VisitorMessage{},
		Page:    page,
		PerPage: 10,
		Total:   0,
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

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool

	c.Status(http.StatusNoContent)
}
