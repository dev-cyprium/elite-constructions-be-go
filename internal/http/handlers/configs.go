package handlers

import (
	"errors"
	"net/http"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if configuration exists
	_, err := queries.GetConfigurationByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "Configuration not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Update configuration
	err = queries.UpdateConfigurationValue(ctx, sqlc.UpdateConfigurationValueParams{
		Key:   key,
		Value: req.Value,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update configuration")
		return
	}

	// Get updated configuration
	updated, err := queries.GetConfigurationByKey(ctx, key)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"key":   updated.Key,
		"value": updated.Value,
	})
}
