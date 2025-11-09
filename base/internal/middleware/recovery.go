package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery middleware recovers from panics and returns 500 error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Printf("PANIC: %v", err)

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred. Please try again later.",
					"code":    "INTERNAL_ERROR",
				})

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}
