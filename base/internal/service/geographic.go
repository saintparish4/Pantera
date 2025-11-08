package service

import (
	"fmt"
	"strings"

	"github.com/saintparish4/harmonia/internal/domain"
)

// GeographicStrategy implements location-based pricing with regional multipliers
type GeographicStrategy struct{}

// Name returns the strategy identifier
func (s *GeographicStrategy) Name() string {
	return domain.StrategyTypeGeographic
}

// Validate checks if the configuration is valid for geographic pricing
func (s *GeographicStrategy) Validate(config map[string]interface{}) error {
	// regional_multipliers is required in config
	multipliers, ok := domain.GetMap(config, "regional_multipliers")
	if !ok || len(multipliers) == 0 {
		return fmt.Errorf("%w: regional_multipliers map is required", domain.ErrConfigurationInvalid)
	}

	// Validate all multipliers are positive numbers
	for region, multiplier := range multipliers {
		var mult float64
		switch v := multiplier.(type) {
		case float64:
			mult = v
		case float32:
			mult = float64(v)
		case int:
			mult = float64(v)
		default:
			return fmt.Errorf("regional_multipliers[%s] must be a number", region)
		}

		if mult <= 0 {
			return fmt.Errorf("regional_multipliers[%s] must be positive", region)
		}
	}

	return nil
}

// Calculate computes the geographic price based on location
func (s *GeographicStrategy) Calculate(req *domain.PricingRequest, config map[string]interface{}) (*domain.PricingResponse, error) {
	// Extract base_price from inputs (required)
	basePrice, ok := domain.GetFloat64(req.Inputs, "base_price")
	if !ok {
		return nil, fmt.Errorf("%w: base_price is required", domain.ErrMissingRequiredField)
	}

	if basePrice < 0 {
		return nil, fmt.Errorf("%w: base_price cannot be negative", domain.ErrInvalidFieldValue)
	}

	// Extract location from inputs (required)
	location, ok := domain.GetString(req.Inputs, "location")
	if !ok || location == "" {
		return nil, fmt.Errorf("%w: location is required", domain.ErrMissingRequiredField)
	}

	// Get regional multipliers from config
	multipliers, _ := domain.GetMap(config, "regional_multipliers")

	// Find the multiplier for this location
	multiplier, regionUsed := findRegionalMultiplier(location, multipliers)

	// Calculate final price
	finalPrice := basePrice * multiplier

	// Apply min/max bounds if specified
	minPrice, _ := domain.GetFloat64(config, "min_price")
	maxPrice, _ := domain.GetFloat64(config, "max_price")
	finalPrice = domain.ApplyBounds(finalPrice, minPrice, maxPrice)

	// Round final price to 2 decimal places for consistency
	finalPrice = domain.RoundToTwoDecimals(finalPrice)

	// Get currency (can be region-specific or default)
	currency := getCurrencyForRegion(location, config)

	// Build response with breakdown
	response := &domain.PricingResponse{
		FinalPrice:    finalPrice,
		OriginalPrice: basePrice,
		Currency:      currency,
		Breakdown: domain.PriceBreakdown{
			BasePrice: basePrice,
			Adjustments: []domain.PriceAdjustment{
				{
					Type:        "regional_multiplier",
					Description: fmt.Sprintf("Regional pricing for %s", regionUsed),
					Amount:      multiplier,
					Applied:     finalPrice - basePrice,
				},
			},
			Details: map[string]interface{}{
				"base_price":        basePrice,
				"location":          location,
				"region_used":       regionUsed,
				"multiplier":        multiplier,
				"final_price":       domain.RoundToTwoDecimals(finalPrice),
				"currency":          currency,
				"available_regions": getAvailableRegions(multipliers),
			},
		},
	}

	return response, nil
}

// findRegionalMultiplier finds the most specific multiplier for a location
// Supports both full location codes (e.g., "US-CA") and country codes (e.g., "US")
func findRegionalMultiplier(location string, multipliers map[string]interface{}) (float64, string) {
	location = strings.ToUpper(strings.TrimSpace(location))

	// Try exact match first (e.g., "US-CA")
	if mult, ok := multipliers[location]; ok {
		if multiplier, ok := convertToFloat(mult); ok {
			return multiplier, location
		}
	}

	// Try country code if location has a dash (e.g., "US-CA" -> "US")
	if strings.Contains(location, "-") {
		countryCode := strings.Split(location, "-")[0]
		if mult, ok := multipliers[countryCode]; ok {
			if multiplier, ok := convertToFloat(mult); ok {
				return multiplier, countryCode
			}
		}
	}

	// Try default multiplier
	if mult, ok := multipliers["default"]; ok {
		if multiplier, ok := convertToFloat(mult); ok {
			return multiplier, "default"
		}
	}

	// No multiplier found, use 1.0 (no change)
	return 1.0, "unknown"
}

// convertToFloat safely converts interface{} to float64
func convertToFloat(val interface{}) (float64, bool) {
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

// getCurrencyForRegion gets currency for a region (can be region-specific or default)
func getCurrencyForRegion(location string, config map[string]interface{}) string {
	location = strings.ToUpper(strings.TrimSpace(location))

	// Check for currency_map in config
	if currencyMap, ok := domain.GetMap(config, "currency_map"); ok {
		// Try exact location match
		if currency, ok := domain.GetString(currencyMap, location); ok {
			return currency
		}

		// Try country code if location has dash
		if strings.Contains(location, "-") {
			countryCode := strings.Split(location, "-")[0]
			if currency, ok := domain.GetString(currencyMap, countryCode); ok {
				return currency
			}
		}
	}

	// Use default currency from config
	if currency, ok := domain.GetString(config, "default_currency"); ok {
		return currency
	}

	return "USD"
}

// getAvailableRegions returns a list of configured regions
func getAvailableRegions(multipliers map[string]interface{}) []string {
	regions := make([]string, 0, len(multipliers))
	for region := range multipliers {
		regions = append(regions, region)
	}
	return regions
}
