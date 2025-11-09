package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyRepository defines the interface for API key validation
type APIKeyRepository interface {
	ValidateKey(ctx context.Context, keyHash string) (userID uuid.UUID, isValid bool, err error)
}

// AuthMiddleware handles API key authentication
type AuthMiddleware struct {
	repo APIKeyRepository
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(repo APIKeyRepository) *AuthMiddleware {
	return &AuthMiddleware{
		repo: repo,
	}
}

// Authenticate validates the API key and sets user context
func (a *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Authorization header as fallback
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(401, gin.H{
				"error":   "Unauthorized",
				"message": "API key is required",
				"code":    "MISSING_API_KEY",
			})
			c.Abort()
			return
		}

		// Validate API key format (basic check)
		if !strings.HasPrefix(apiKey, "hm_live_") && !strings.HasPrefix(apiKey, "hm_test_") {
			c.JSON(401, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid API key format",
				"code":    "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		// Validate API key with repository
		ctx := c.Request.Context()
		userID, isValid, err := a.repo.ValidateKey(ctx, apiKey)
		if err != nil {
			c.JSON(500, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to validate API key",
				"code":    "VALIDATION_ERROR",
			})
			c.Abort()
			return
		}

		if !isValid {
			c.JSON(401, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or revoked API key",
				"code":    "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		// Set user ID in context for downstream handlers
		c.Set("user_id", userID)
		c.Set("api_key", apiKey)

		c.Next()
	}
}

// OptionalAuth is like Authenticate but doesn't abort on missing/invalid keys
// Useful for endpoints that work differently for authenticated vs anonymous users
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey != "" && (strings.HasPrefix(apiKey, "hm_live_") || strings.HasPrefix(apiKey, "hm_test_")) {
			ctx := c.Request.Context()
			userID, isValid, err := a.repo.ValidateKey(ctx, apiKey)
			if err == nil && isValid {
				c.Set("user_id", userID)
				c.Set("api_key", apiKey)
			}
		}

		c.Next()
	}
}
