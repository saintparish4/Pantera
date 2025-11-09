package dto

import (
	"time"

	"github.com/google/uuid"
)

// --- Pricing Calculation DTOs ---

// CalculatePriceRequest represents a pricing calculation request
type CalculatePriceRequest struct {
	StrategyType string                 `json:"strategy_type" binding:"required"`
	BasePrice    float64                `json:"base_price" binding:"required,gt=0"`
	Quantity     int                    `json:"quantity" binding:"required,gt=0"`
	Context      map[string]interface{} `json:"context,omitempty"`
	ProductSKU   string                 `json:"product_sku,omitempty"`
}

// CalculatePriceResponse represents the pricing calculation result
type CalculatePriceResponse struct {
	FinalPrice   float64                `json:"final_price"`
	StrategyType string                 `json:"strategy_type"`
	Breakdown    map[string]interface{} `json:"breakdown"`
	CalculatedAt time.Time              `json:"calculated_at"`
}

// --- Pricing Strategy DTOs ---

// PricingStrategyResponse represents a pricing strategy
type PricingStrategyResponse struct {
	Type           string   `json:"type"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	RequiredFields []string `json:"required_fields"`
}

// --- Pricing Rule DTOs ---

// CreatePricingRuleRequest represents a request to create a pricing rule
type CreatePricingRuleRequest struct {
	Name         string                 `json:"name" binding:"required"`
	StrategyType string                 `json:"strategy_type" binding:"required"`
	Config       map[string]interface{} `json:"config" binding:"required"`
	IsActive     bool                   `json:"is_active"`
}

// UpdatePricingRuleRequest represents a request to update a pricing rule
type UpdatePricingRuleRequest struct {
	Name         *string                `json:"name,omitempty"`
	StrategyType *string                `json:"strategy_type,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
}

// PricingRuleResponse represents a pricing rule
type PricingRuleResponse struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`
	Name         string                 `json:"name"`
	StrategyType string                 `json:"strategy_type"`
	Config       map[string]interface{} `json:"config"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// --- Product DTOs ---

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	SKU         string  `json:"sku" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description,omitempty"`
	BaseCost    float64 `json:"base_cost" binding:"required,gte=0"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	BaseCost    *float64 `json:"base_cost,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// ProductResponse represents a product
type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	BaseCost    float64   `json:"base_cost"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- Calculation Log DTOs ---

// CalculationLogResponse represents a calculation log entry
type CalculationLogResponse struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`
	StrategyType string                 `json:"strategy_type"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	CreatedAt    time.Time              `json:"created_at"`
}

// LogsQueryParams represents query parameters for fetching logs
type LogsQueryParams struct {
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset       int    `form:"offset" binding:"omitempty,min=0"`
	StrategyType string `form:"strategy_type"`
	StartDate    string `form:"start_date"` // ISO 8601 format
	EndDate      string `form:"end_date"`   // ISO 8601 format
}

// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}
