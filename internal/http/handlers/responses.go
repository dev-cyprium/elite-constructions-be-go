package handlers

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, details ...interface{}) {
	response := gin.H{
		"error": message,
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	c.JSON(statusCode, response)
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}
