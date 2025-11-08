package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Pricing Strategy Types
const (
	StrategyTypeCostPlus   = "cost_plus"
	StrategyTypeGeographic = "geographic"
	StrategyTypeTimeBased  = "time_based"
	StrategyTypeRuleBased  = "rule_based"
)

// Pricing Strategy defines the interface all pricing strategies must implement
type PricingStrategy interface {
	// Calculate computes the price based on the reqeust and rule config
	Calculate(req *PricingRequest, config map[string]interface{}) (*PricingResponse, error)

	// Validate checks if the config si valid for this strategy
	Validate(config map[string]interface{}) error

	// Name returns the strategy type identifier
	Name() string
}

// PricingRequest contains all inputs for price calculation
type PricingRequest struct {
	// Strategy Selection
	Strategy string `json:"strategy" binding:"required"`

	// Optional: Use a saved rule instead of inline config
	RuleID *uuid.UUID `json:"rule_id,omitempty"`

	// Optional: Product SKU for product-based pricing
	ProductSKU string `json:"product_sku,omitempty"`

	// Strategy specific inputs (flexible map for all strategies)
	Inputs map[string]interface{} `json:"inputs" binding:"required"`

	// Metadata for audit trail
	RequestID   string    `json:"request_id,omitempty"`
	RequestedAt time.Time `json:"requested_at,omitempty"`
}

// PricingResponse contains the calculated price and detailed breakdown
type PricingResponse struct {
	// Core pricing data
	FinalPrice    float64 `json:"final_price"`
	OriginalPrice float64 `json:"original_price,omitempty"`
	Currency      string  `json:"currency"`

	// Breakdown of how price was calculated
	Breakdown PriceBreakdown `json:"breakdown"`

	// Metadata
	Strategy      string                 `json:"strategy"`
	AppliedRuleID *uuid.UUID             `json:"applied_rule_id,omitempty"`
	CalculatedAt  time.Time              `json:"calculated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PriceBreakdown provides transparency into price calculation
type PriceBreakdown struct {
	// Common fields
	BasePrice   float64           `json:"base_price,omitempty"`
	Adjustments []PriceAdjustment `json:"adjustments,omitempty"`

	// Strategy specific details (flexible map)
	Details map[string]interface{} `json:"details,omitempty"`
}

// PriceAdjustment represents a single modification to the price
type PriceAdjustment struct {
	Type        string  `json:"type"`        // "markup", ",multiplier", "tax", "discount", "fee", "other"
	Description string  `json:"description"` // Human-readable explanation
	Amount      float64 `json:"amount"`      // Dollar amount or multiplier
	Applied     float64 `json:"applied"`     // Actual value applied
}

// PricingRule represents a saved pricing config
type PricingRule struct {
	ID           uuid.UUID              `json:"id"`
	UserID       uuid.UUID              `json:"user_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	StrategyType string                 `json:"strategy_type"`
	Config       map[string]interface{} `json:"config"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty"`
}

// Product represents a SKU in the catalog
type Product struct {
	ID            uuid.UUID              `json:"id"`
	UserID        uuid.UUID              `json:"user_id"`
	SKU           string                 `json:"sku"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	BaseCost      float64                `json:"base_cost"`
	DefaultRuleID *uuid.UUID             `json:"default_rule_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	IsActive      bool                   `json:"is_active"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CalculationLog represents an audit trail entry
type CalculationLog struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	APIKeyID        *uuid.UUID             `json:"api_key_id,omitempty"`
	RuleID          *uuid.UUID             `json:"rule_id,omitempty"`
	StrategyType    string                 `json:"strategy_type"`
	InputData       map[string]interface{} `json:"input_data"`
	OutputData      map[string]interface{} `json:"output_data"`
	ExecutionTimeMs int                    `json:"execution_time_ms"`
	CreatedAt       time.Time              `json:"created_at"`
}

// Common validation errors
var (
	ErrInvalidStrategy      = errors.New("invalid pricing strategy")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidFieldValue    = errors.New("invalid field value")
	ErrConfigurationInvalid = errors.New("strategy configuration is invalid")
	ErrNegativePrice        = errors.New("price cannot be negative")
	ErrPriceExceedsLimits   = errors.New("calculated price exceeds maximum allowed")
	ErrPriceBelowMin        = errors.New("calculated price is below minimum allowed")
)

// ValidateStrategy checks if a strategy type is supported
func ValidateStrategy(strategy string) error {
	validStrategies := map[string]bool{
		StrategyTypeCostPlus:   true,
		StrategyTypeGeographic: true,
		StrategyTypeRuleBased:  true,
		StrategyTypeTimeBased:  true,
	}

	if !validStrategies[strategy] {
		return ErrInvalidStrategy
	}
	return nil
}

// GetFloat64 safely extracts a float64 from an interface map
func GetFloat64(m map[string]interface{}, key string) (float64, bool) {
	val, exists := m[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetString safely extracts a string from an interface map
func GetString(m map[string]interface{}, key string) (string, bool) {
	val, exists := m[key]
	if !exists {
		return "", false
	}

	if str, ok := val.(string); ok {
		return str, true
	}

	return "", false
}

// GetMap safely extracts a map from an interface map
func GetMap(m map[string]interface{}, key string) (map[string]interface{}, bool) {
	val, exists := m[key]
	if !exists {
		return nil, false
	}

	if mapVal, ok := val.(map[string]interface{}); ok {
		return mapVal, true
	}

	return nil, false
}

// RoundToTwoDecimals rounds a float to 2 decimal places
func RoundToTwoDecimals(val float64) float64 {
	return float64(int(val*100+0.5)) / 100
}

// ApplyBounds ensures price is within min/max constraints
func ApplyBounds(price, min, max float64) float64 {
	if min > 0 && price < min {
		return min
	}
	if max > 0 && price > max {
		return max
	}
	return price
}
