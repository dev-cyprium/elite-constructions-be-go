package middleware

import (
	"log"
	"net/http"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				handlers.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
