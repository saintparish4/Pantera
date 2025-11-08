package service

import (
	"testing"

	"github.com/saintparish4/harmonia/internal/domain"
)

func TestGeographicStrategy_Calculate(t *testing.T) {
	strategy := &GeographicStrategy{}

	tests := []struct {
		name        string
		request     *domain.PricingRequest
		config      map[string]interface{}
		wantPrice   float64
		wantErr     bool
		errContains string
	}{
		{
			name: "exact region match",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
					"location":   "US-CA",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US":    1.0,
					"US-CA": 1.15,
					"US-NY": 1.20,
				},
			},
			wantPrice: 115.0, // 100 * 1.15
			wantErr:   false,
		},
		{
			name: "country fallback",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
					"location":   "US-TX",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US":    1.05,
					"US-CA": 1.15,
				},
			},
			wantPrice: 105.0, // Falls back to US country code
			wantErr:   false,
		},
		{
			name: "default multiplier",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
					"location":   "ZZ-UNKNOWN",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US":      1.0,
					"default": 0.9,
				},
			},
			wantPrice: 90.0, // Uses default multiplier
			wantErr:   false,
		},
		{
			name: "no multiplier found",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
					"location":   "XX",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US": 1.0,
				},
			},
			wantPrice: 100.0, // No match, uses 1.0
			wantErr:   false,
		},
		{
			name: "missing base_price",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"location": "US-CA",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US-CA": 1.15,
				},
			},
			wantErr:     true,
			errContains: "base_price",
		},
		{
			name: "missing location",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US": 1.0,
				},
			},
			wantErr:     true,
			errContains: "location",
		},
		{
			name: "case insensitive location",
			request: &domain.PricingRequest{
				Strategy: domain.StrategyTypeGeographic,
				Inputs: map[string]interface{}{
					"base_price": 100.0,
					"location":   "us-ca",
				},
			},
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US-CA": 1.15,
				},
			},
			wantPrice: 115.0, // Should match despite case difference
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

			// Verify location in breakdown
			if details, ok := response.Breakdown.Details["location"]; !ok {
				t.Error("breakdown should include location")
			} else if details == "" {
				t.Error("location should not be empty")
			}
		})
	}
}

func TestGeographicStrategy_Validate(t *testing.T) {
	strategy := &GeographicStrategy{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US":    1.0,
					"US-CA": 1.15,
				},
			},
			wantErr: false,
		},
		{
			name:    "missing regional_multipliers",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty regional_multipliers",
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "invalid multiplier type",
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US": "not_a_number",
				},
			},
			wantErr: true,
		},
		{
			name: "negative multiplier",
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US": -1.0,
				},
			},
			wantErr: true,
		},
		{
			name: "zero multiplier",
			config: map[string]interface{}{
				"regional_multipliers": map[string]interface{}{
					"US": 0.0,
				},
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