package service

import (
	"testing"

	"github.com/saintparish4/harmonia/internal/domain"
)

func TestPricingEngine_Calculate(t *testing.T) {
	engine := NewPricingEngine()

	tests := []struct {
		name        string
		request     *domain.PricingRequest
		config      map[string]interface{}
		wantErr     bool
		errContains string
		checkPrice  bool
		minPrice    float64
	}{
		{
			name: "cost_plus strategy",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    100.0,
					"markup_value": 25.0,
				},
			},
			config: map[string]interface{}{
				"markup_type": "percentage",
			},
			wantErr:    false,
			checkPrice: true,
			minPrice:   125.0,
		},
		{
			name: "geographic strategy",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
					"location":   "US-CA",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US-CA": 1.15,
				},
			},
			wantErr:    false,
			checkPrice: true,
			minPrice:   115.0,
		},
		{
			name: "invalid strategy",
			request: &domain.PricingRequest{
				Strategy: "invalid_strategy",
				Inputs:   map[string]interface{}{},
			},
			config:      map[string]interface{}{},
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name: "validation fails",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"markup_value": 25.0,
					// Missing base_cost
				},
			},
			config:      map[string]interface{}{},
			wantErr:     true,
			errContains: "base_cost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := engine.Calculate(tt.request, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Fatal("response should not be nil")
			}

			// Check strategy is set
			if response.Strategy != tt.request.Strategy {
				t.Errorf("expected strategy %s, got %s", tt.request.Strategy, response.Strategy)
			}

			// Check calculated_at is set
			if response.CalculatedAt.IsZero() {
				t.Error("calculated_at should be set")
			}

			// Check price is positive
			if response.FinalPrice < 0 {
				t.Error("final_price should not be negative")
			}

			// Check minimum price if specified
			if tt.checkPrice && response.FinalPrice < tt.minPrice {
				t.Errorf("expected price >= %.2f, got %.2f", tt.minPrice, response.FinalPrice)
			}

			// Check breakdown exists
			if len(response.Breakdown.Adjustments) == 0 && response.Strategy != "rule_based" {
				t.Error("breakdown should have at least one adjustment")
			}
		})
	}
}

func TestPricingEngine_ListStrategies(t *testing.T) {
	engine := NewPricingEngine()

	strategies := engine.ListStrategies()

	if len(strategies) != 4 {
		t.Errorf("expected 4 strategies, got %d", len(strategies))
	}

	// Check all expected strategies are present
	expectedStrategies := map[string]bool{
		domain.StrategyTypeCostPlus:   false,
		domain.StrategyTypeGeographic: false,
		domain.StrategyTypeTimeBased:  false,
		domain.StrategyTypeRuleBased:  false,
	}

	for _, strategy := range strategies {
		if _, ok := expectedStrategies[strategy]; ok {
			expectedStrategies[strategy] = true
		}
	}

	for strategy, found := range expectedStrategies {
		if !found {
			t.Errorf("strategy %s not found in list", strategy)
		}
	}
}

func TestPricingEngine_GetStrategy(t *testing.T) {
	engine := NewPricingEngine()

	tests := []struct {
		name         string
		strategyName string
		wantErr      bool
	}{
		{
			name:         "get cost_plus",
			strategyName: domain.StrategyTypeCostPlus,
			wantErr:      false,
		},
		{
			name:         "get invalid strategy",
			strategyName: "invalid",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := engine.GetStrategy(tt.strategyName)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if strategy == nil {
				t.Error("strategy should not be nil")
			}

			if strategy.Name() != tt.strategyName {
				t.Errorf("expected strategy name %s, got %s", tt.strategyName, strategy.Name())
			}
		})
	}
}

func TestPricingEngine_ValidateConfig(t *testing.T) {
	engine := NewPricingEngine()

	tests := []struct {
		name         string
		strategyName string
		config       map[string]interface{}
		wantErr      bool
	}{
		{
			name:         "valid cost_plus config",
			strategyName: domain.StrategyTypeCostPlus,
			config: map[string]interface{}{
				"markup_value": 25.0,
			},
			wantErr: false,
		},
		{
			name:         "invalid cost_plus config",
			strategyName: domain.StrategyTypeCostPlus,
			config: map[string]interface{}{
				"markup_value": -10.0,
			},
			wantErr: true,
		},
		{
			name:         "valid geographic config",
			strategyName: domain.StrategyTypeGeographic,
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US": 1.0,
				},
			},
			wantErr: false,
		},
		{
			name:         "invalid geographic config - no multipliers",
			strategyName: domain.StrategyTypeGeographic,
			config:       map[string]interface{}{},
			wantErr:      true,
		},
		{
			name:         "invalid strategy name",
			strategyName: "invalid",
			config:       map[string]interface{}{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateConfig(tt.strategyName, tt.config)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}