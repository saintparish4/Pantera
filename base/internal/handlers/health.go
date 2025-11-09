package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker defines interface for database health checks
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles health check endpoint
type HealthHandler struct {
	db HealthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db HealthChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check handles GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	dbStatus := "healthy"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	status := "healthy"
	if dbStatus == "unhealthy" {
		status = "degraded"
	}

	response := gin.H{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"services": gin.H{
			"database": dbStatus,
		},
	}

	Success(c, response)
}
