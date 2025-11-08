package service

import (
	"testing"

	"github.com/saintparish4/harmonia/internal/domain"
)

func TestCostPlusStrategy_Calculate(t *testing.T) {
	strategy := &CostPlusStrategy{}

	tests := []struct {
		name        string
		request     *domain.PricingRequest
		config      map[string]interface{}
		wantPrice   float64
		wantErr     bool
		errContains string
	}{
		{
			name: "basic percentage markup",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    100.0,
					"markup_type":  "percentage",
					"markup_value": 25.0,
				},
			},
			config:    map[string]interface{}{},
			wantPrice: 125.0,
			wantErr:   false,
		},
		{
			name: "percentage markup with tax",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    100.0,
					"markup_type":  "percentage",
					"markup_value": 25.0,
					"tax_rate":     0.08,
				},
			},
			config:    map[string]interface{}{},
			wantPrice: 135.0, // 100 + 25 = 125, + 8% tax = 135
			wantErr:   false,
		},
		{
			name: "fixed markup",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    100.0,
					"markup_type":  "fixed",
					"markup_value": 30.0,
				},
			},
			config:    map[string]interface{}{},
			wantPrice: 130.0,
			wantErr:   false,
		},
		{
			name: "markup from config",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost": 100.0,
				},
			},
			config: map[string]interface{}{
				"markup_type":  "percentage",
				"markup_value": 50.0,
			},
			wantPrice: 150.0,
			wantErr:   false,
		},
		{
			name: "missing base_cost",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"markup_value": 25.0,
				},
			},
			config:      map[string]interface{}{},
			wantErr:     true,
			errContains: "base_cost",
		},
		{
			name: "negative base_cost",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    -100.0,
					"markup_value": 25.0,
				},
			},
			config:      map[string]interface{}{},
			wantErr:     true,
			errContains: "negative",
		},
		{
			name: "apply min price",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    10.0,
					"markup_type":  "percentage",
					"markup_value": 10.0,
				},
			},
			config: map[string]interface{}{
				"min_price": 50.0,
			},
			wantPrice: 50.0, // Would be 11, but min is 50
			wantErr:   false,
		},
		{
			name: "apply max price",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeCostPlus,
				Inputs: map[string]interface{}{
					"base_cost":    100.0,
					"markup_type":  "percentage",
					"markup_value": 100.0,
				},
			},
			config: map[string]interface{}{
				"max_price": 150.0,
			},
			wantPrice: 150.0, // Would be 200, but max is 150
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := strategy.Calculate(tt.request, tt.config)

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

			if response.FinalPrice != tt.wantPrice {
				t.Errorf("expected price %.2f, got %.2f", tt.wantPrice, response.FinalPrice)
			}

			// Verify breakdown exists
			if response.Breakdown.BasePrice == 0 {
				t.Error("breakdown base_price should not be zero")
			}
		})
	}
}

func TestCostPlusStrategy_Validate(t *testing.T) {
	strategy := &CostPlusStrategy{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty config is valid",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "valid markup value",
			config: map[string]interface{}{
				"markup_value": 25.0,
			},
			wantErr: false,
		},
		{
			name: "negative markup value",
			config: map[string]interface{}{
				"markup_value": -10.0,
			},
			wantErr: true,
		},
		{
			name: "valid tax rate",
			config: map[string]interface{}{
				"tax_rate": 0.08,
			},
			wantErr: false,
		},
		{
			name: "invalid tax rate - negative",
			config: map[string]interface{}{
				"tax_rate": -0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid tax rate - over 1",
			config: map[string]interface{}{
				"tax_rate": 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.Validate(tt.config)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}