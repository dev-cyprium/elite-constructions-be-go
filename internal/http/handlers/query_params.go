package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// QueryParams represents parsed query parameters for list endpoints
type QueryParams struct {
	Search    string
	SortBy    string
	SortOrder string
}

// ParseQueryParams extracts common query parameters from gin.Context
// This can be reused across different endpoints
func ParseQueryParams(c *gin.Context) QueryParams {
	return QueryParams{
		Search:    strings.TrimSpace(c.DefaultQuery("search", "")),
		SortBy:    strings.TrimSpace(c.DefaultQuery("sort_by", "")),
		SortOrder: strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_order", "asc"))),
	}
}

// ValidateSortOrder validates and normalizes sort order (asc/desc)
// Returns "asc" if invalid or empty
func ValidateSortOrder(sortOrder string) string {
	sortOrder = strings.ToLower(strings.TrimSpace(sortOrder))
	if sortOrder != "asc" && sortOrder != "desc" {
		return "asc"
	}
	return sortOrder
}

// ValidateSortBy validates if the sort_by field is allowed
// Returns the sort_by if valid, empty string otherwise
func ValidateSortBy(sortBy string, allowedFields []string) string {
	sortBy = strings.TrimSpace(sortBy)
	for _, field := range allowedFields {
		if sortBy == field {
			return sortBy
		}
	}
	return ""
}
