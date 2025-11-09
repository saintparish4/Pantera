package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/dto"
)

// CalculationLog represents a calculation log entry
type CalculationLog struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	StrategyType string
	Input        map[string]interface{}
	Output       map[string]interface{}
	CreatedAt    time.Time
}

// CalculationLogRepository defines operations for calculation logs
type CalculationLogRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, strategyType string, startDate, endDate *time.Time) ([]*CalculationLog, error)
	Count(ctx context.Context, userID uuid.UUID, strategyType string, startDate, endDate *time.Time) (int, error)
}

// LogsHandler handles calculation log endpoints
type LogsHandler struct {
	repo CalculationLogRepository
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(repo CalculationLogRepository) *LogsHandler {
	return &LogsHandler{repo: repo}
}

// List handles GET /v1/logs
func (h *LogsHandler) List(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Bind query parameters
	var params dto.LogsQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	// Parse date filters if provided
	var startDate, endDate *time.Time
	if params.StartDate != "" {
		t, err := time.Parse(time.RFC3339, params.StartDate)
		if err != nil {
			BadRequest(c, "Invalid start_date format. Use ISO 8601 (RFC3339)")
			return
		}
		startDate = &t
	}
	if params.EndDate != "" {
		t, err := time.Parse(time.RFC3339, params.EndDate)
		if err != nil {
			BadRequest(c, "Invalid end_date format. Use ISO 8601 (RFC3339)")
			return
		}
		endDate = &t
	}

	ctx := c.Request.Context()

	// Get logs
	logs, err := h.repo.GetByUserID(ctx, userID, params.Limit, params.Offset, params.StrategyType, startDate, endDate)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Get total count
	total, err := h.repo.Count(ctx, userID, params.StrategyType, startDate, endDate)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Convert to response DTOs
	logResponses := make([]dto.CalculationLogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = dto.CalculationLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			StrategyType: log.StrategyType,
			Input:        log.Input,
			Output:       log.Output,
			CreatedAt:    log.CreatedAt,
		}
	}

	// Create paginated response
	response := dto.PaginatedResponse{
		Data:    logResponses,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: params.Offset+params.Limit < total,
	}

	Success(c, response)
}
