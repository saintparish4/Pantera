package models

import (
	"encoding/json"
	"time"
)

type PricingRule struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Strategy         string    `json:"strategy"` // cost_plus, demand_based, competitive
	BasePrice        float64   `json:"base_price"`
	MarkupPercentage float64   `json:"markup_percentage"`
	MinPrice         float64   `json:"min_price"`
	MaxPrice         float64   `json:"max_price"`
	DemandMultiplier float64   `json:"demand_multiplier"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PriceRequest struct {
	RuleID          int                    `json:"rule_id" binding:"required"`
	Quantity        int                    `json:"quantity"`
	DemandLevel     int                    `json:"demand_level"` // 0.0 to 2.0 (low, medium, high)
	CompetitorPrice float64                `json:"competitor_price"`
	CustomData      map[string]interface{} `json:"custom_data"` // Additional data for custom logic
}

type PriceResponse struct {
	Price         float64                `json:"price"`
	OriginalPrice float64                `json:"original_price"`
	Strategy      string                 `json:"strategy"`
	Breakdown     map[string]interface{} `json:"breakdown"`
	CalculatedAt  time.Time              `json:"calculated_at"`
}

type PriceCalculation struct {
	ID              int             `json:"id"`
	RuleID          int             `json:"rule_id"`
	InputData       json.RawMessage `json:"input_data"`
	CalculatedPrice float64         `json:"calculated_price"`
	StrategyUsed    string          `json:"strategy_used"`
	CreatedAt       time.Time       `json:"created_at"`
}

type CreateRuleRequest struct {
	Name             string  `json:"name" binding:"required"`
	Strategy         string  `json:"strategy" binding:"required"`
	BasePrice        float64 `json:"base_price" binding:"required"`
	MarkupPercentage float64 `json:"markup_percentage"`
	MinPrice         float64 `json:"min_price"`
	MaxPrice         float64 `json:"max_price"`
	DemandMultiplier float64 `json:"demand_multiplier"`
}
