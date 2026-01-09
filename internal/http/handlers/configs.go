package handlers

import (
	"net/http"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/gin-gonic/gin"
)

type UpdateConfigRequest struct {
	Value string `json:"value" binding:"required"`
}

// UpdateConfig updates a configuration value by key
func UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		ErrorResponse(c, http.StatusBadRequest, "Configuration key is required")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// TODO: Implement with sqlc queries
	_ = db.Pool
	_ = key
	_ = req.Value

	SuccessResponse(c, http.StatusOK, gin.H{
		"key":   key,
		"value": req.Value,
	})
}
