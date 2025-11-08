package service

import (
	"fmt"
	"time"

	"github.com/saintparish4/harmonia/internal/domain"
)

// PricingEngine coordinates all pricing strategies
type PricingEngine struct {
	strategies map[string]domain.PricingStrategy
}

// NewPricingEngine creates a new pricing engine with all strategies registered
func NewPricingEngine() *PricingEngine {
	engine := &PricingEngine{
		strategies: make(map[string]domain.PricingStrategy),
	}

	// Register all pricing strategies
	engine.RegisterStrategy(&CostPlusStrategy{})
	engine.RegisterStrategy(&GeographicStrategy{})
	engine.RegisterStrategy(&TimeBasedStrategy{})
	engine.RegisterStrategy(&RuleBasedStrategy{})

	return engine
}

// RegisterStrategy adds a pricing strategy to the engine
func (e *PricingEngine) RegisterStrategy(strategy domain.PricingStrategy) {
	e.strategies[strategy.Name()] = strategy
}

// Calculate processes a pricing request and returns the calculated price
func (e *PricingEngine) Calculate(req *domain.PricingRequest, config map[string]interface{}) (*domain.PricingResponse, error) {
	// Validate strategy exists
	if err := domain.ValidateStrategy(req.Strategy); err != nil {
		return nil, err
	}

	// Get the appropriate strategy
	strategy, exists := e.strategies[req.Strategy]
	if !exists {
		return nil, fmt.Errorf("strategy not registered: %s", req.Strategy)
	}

	// Validate configuration for this strategy
	if err := strategy.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set request timestamp if not provided
	if req.RequestedAt.IsZero() {
		req.RequestedAt = time.Now()
	}

	// Calculate price using the strategy
	response, err := strategy.Calculate(req, config)
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	// Set response metadata
	response.Strategy = req.Strategy
	response.CalculatedAt = time.Now()
	response.AppliedRuleID = req.RuleID

	// Round final price to 2 decimals
	response.FinalPrice = domain.RoundToTwoDecimals(response.FinalPrice)

	// Validate final price
	if response.FinalPrice < 0 {
		return nil, domain.ErrNegativePrice
	}

	return response, nil
}

// GetStrategy returns a registered strategy by name
func (e *PricingEngine) GetStrategy(name string) (domain.PricingStrategy, error) {
	strategy, exists := e.strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy not found: %s", name)
	}
	return strategy, nil
}

// ListStrategies returns all registered strategy names
func (e *PricingEngine) ListStrategies() []string {
	strategies := make([]string, 0, len(e.strategies))
	for name := range e.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

// ValidateConfig validates a configuration for a specific strategy
func (e *PricingEngine) ValidateConfig(strategyName string, config map[string]interface{}) error {
	strategy, err := e.GetStrategy(strategyName)
	if err != nil {
		return err
	}
	return strategy.Validate(config)
}
