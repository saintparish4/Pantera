package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saintparish4/harmonia/internal/dto"
)

// PricingRequest represents a pricing calculation request
type PricingRequest struct {
	StrategyType string
	BasePrice    float64
	Quantity     int
	Context      map[string]interface{}
}

// PricingResult represents a pricing calculation result
type PricingResult struct {
	FinalPrice float64
	Breakdown  map[string]interface{}
}

// PricingEngine defines interface for pricing calculations
type PricingEngine interface {
	Calculate(ctx context.Context, req *PricingRequest) (*PricingResult, error)
	GetAvailableStrategies() []string
}

// CalculationLogger defines interface for logging calculations
type CalculationLogger interface {
	Log(ctx context.Context, userID uuid.UUID, strategyType string, input, output map[string]interface{}) error
}

// PricingHandler handles pricing-related endpoints
type PricingHandler struct {
	engine PricingEngine
	logger CalculationLogger
}

// NewPricingHandler creates a new pricing handler
func NewPricingHandler(engine PricingEngine, logger CalculationLogger) *PricingHandler {
	return &PricingHandler{
		engine: engine,
		logger: logger,
	}
}

// ListStrategies handles GET /v1/pricing/strategies
func (h *PricingHandler) ListStrategies(c *gin.Context) {
	strategies := []dto.PricingStrategyResponse{
		{
			Type:           "cost_plus",
			Name:           "Cost-Plus Pricing",
			Description:    "Adds a fixed markup or percentage to the base cost",
			RequiredFields: []string{"base_price", "markup_type", "markup_value"},
		},
		{
			Type:           "geographic",
			Name:           "Geographic Pricing",
			Description:    "Applies regional multipliers based on location",
			RequiredFields: []string{"base_price", "region_code"},
		},
		{
			Type:           "time_based",
			Name:           "Time-Based Surge Pricing",
			Description:    "Adjusts price based on time windows and demand",
			RequiredFields: []string{"base_price"},
		},
		{
			Type:           "rule_based",
			Name:           "Rule-Based Pricing",
			Description:    "Applies custom conditional rules to determine price",
			RequiredFields: []string{"base_price", "context"},
		},
	}

	Success(c, strategies)
}

// Calculate handles POST /v1/pricing/calculate
func (h *PricingHandler) Calculate(c *gin.Context) {
	// Get user ID from context
	userID := MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	// Bind request
	var req dto.CalculatePriceRequest
	if !BindJSON(c, &req) {
		return
	}

	// Validate strategy type
	validStrategies := map[string]bool{
		"cost_plus":  true,
		"geographic": true,
		"time_based": true,
		"rule_based": true,
	}

	if !validStrategies[req.StrategyType] {
		BadRequest(c, "Invalid strategy type. Must be one of: cost_plus, geographic, time_based, rule_based")
		return
	}

	ctx := c.Request.Context()

	// Prepare pricing request
	pricingReq := &PricingRequest{
		StrategyType: req.StrategyType,
		BasePrice:    req.BasePrice,
		Quantity:     req.Quantity,
		Context:      req.Context,
	}

	// Calculate price
	result, err := h.engine.Calculate(ctx, pricingReq)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Log calculation
	inputData := map[string]interface{}{
		"strategy_type": req.StrategyType,
		"base_price":    req.BasePrice,
		"quantity":      req.Quantity,
		"context":       req.Context,
		"product_sku":   req.ProductSKU,
	}

	outputData := map[string]interface{}{
		"final_price": result.FinalPrice,
		"breakdown":   result.Breakdown,
	}

	go func() {
		// Log asynchronously to not block response
		bgCtx := context.Background()
		if err := h.logger.Log(bgCtx, userID, req.StrategyType, inputData, outputData); err != nil {
			// Log error but don't fail the request
			// In production, you'd want proper logging here
		}
	}()

	// Return response
	response := dto.CalculatePriceResponse{
		FinalPrice:   result.FinalPrice,
		StrategyType: req.StrategyType,
		Breakdown:    result.Breakdown,
		CalculatedAt: time.Now().UTC(),
	}

	Success(c, response)
}
