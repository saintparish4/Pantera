package service

import (
	"fmt"

	"github.com/saintparish4/harmonia/internal/domain"
)

// CostPlusStrategy implements simple markup pricing
type CostPlusStrategy struct{}

// Name returns the strategy identifier
func (s *CostPlusStrategy) Name() string {
	return domain.StrategyTypeCostPlus
}

// Validate checks if the config is valid for cost plus pricing
func (s *CostPlusStrategy) Validate(config map[string]interface{}) error {
	// No required config fields - all inputs come from request
	// Optional config: markup_type, markup_value, tax_rate, min_price, max_price

	// If markup values are in config, validate them
	if markupValue, ok := domain.GetFloat64(config, "markup_value"); ok {
		if markupValue < 0 {
			return fmt.Errorf("markup_value cannot be negative")
		}
	}

	if taxRate, ok := domain.GetFloat64(config, "tax_rate"); ok {
		if taxRate < 0 || taxRate > 1 {
			return fmt.Errorf("tax_rate must be between 0 and 1")
		}
	}
	return nil
}

// Calculate computes the cost-plus price
func (s *CostPlusStrategy) Calculate(req *domain.PricingRequest, config map[string]interface{}) (*domain.PricingResponse, error) {
	// Extract base_cost from inputs (required)
	baseCost, ok := domain.GetFloat64(req.Inputs, "base_cost")
	if !ok {
		return nil, fmt.Errorf("%w: base_cost is required", domain.ErrMissingRequiredField)
	}

	if baseCost < 0 {
		return nil, fmt.Errorf("%w: base_cost cannot be negative", domain.ErrInvalidFieldValue)
	}

	// Get markup_type (default: "percentage")
	markupType, _ := domain.GetString(req.Inputs, "markup_type")
	if markupType == "" {
		markupType, _ = domain.GetString(config, "markup_type")
		if markupType == "" {
			markupType = "percentage"
		}
	}

	// Get markup_value (required, can be inputs or config)
	markupValue, ok := domain.GetFloat64(req.Inputs, "markup_value")
	if !ok {
		markupValue, ok = domain.GetFloat64(config, "markup_value")
		if !ok {
			return nil, fmt.Errorf("%w: markup_value is required", domain.ErrMissingRequiredField)
		}
	}

	// Get optional tax_rate
	taxRate, _ := domain.GetFloat64(req.Inputs, "tax_rate")
	if taxRate == 0 {
		taxRate, _ = domain.GetFloat64(config, "tax_rate")
	}

	// Calculate markup
	var markupAmount float64
	var subtotal float64

	switch markupType {
	case "percentage":
		markupAmount = baseCost * (markupValue / 100)
		subtotal = baseCost + markupAmount
	case "fixed":
		markupAmount = markupValue
		subtotal = baseCost + markupAmount
	default:
		return nil, fmt.Errorf("%w: invalid markup_type: %s", domain.ErrInvalidFieldValue, markupType)
	}

	// Calculate tax
	taxAmount := subtotal * taxRate
	finalPrice := subtotal + taxAmount

	// Apply min/,ax bounds if specified in config
	minPrice, _ := domain.GetFloat64(config, "min_price")
	maxPrice, _ := domain.GetFloat64(config, "max_price")
	finalPrice = domain.ApplyBounds(finalPrice, minPrice, maxPrice)

	// Build response with breakdown
	response := &domain.PricingResponse{
		FinalPrice:    finalPrice,
		OriginalPrice: baseCost,
		Currency:      getCurrency(req, config),
		Breakdown: domain.PriceBreakdown{
			BasePrice: baseCost,
			Adjustments: []domain.PriceAdjustment{
				{
					Type:        "markup",
					Description: fmt.Sprintf("%s markup", markupType),
					Amount:      markupValue,
					Applied:     markupAmount,
				},
			},
			Details: map[string]interface{}{
				"base_cost":     baseCost,
				"markup_type":   markupType,
				"markup_value":  markupValue,
				"markup_amount": domain.RoundToTwoDecimals(markupAmount),
				"subtotal":      domain.RoundToTwoDecimals(subtotal),
				"tax_rate":      taxRate,
				"tax_amount":    domain.RoundToTwoDecimals(taxAmount),
				"final_price":   domain.RoundToTwoDecimals(finalPrice),
			},
		},
	}

	// Add tax adjustment if applicable
	if taxAmount > 0 {
		response.Breakdown.Adjustments = append(response.Breakdown.Adjustments, domain.PriceAdjustment{
			Type:        "tax",
			Description: "sales tax",
			Amount:      taxRate,
			Applied:     taxAmount,
		})
	}
	return response, nil
}

// getCurrency extracts currency from inputs or config (default: USD)
func getCurrency(req *domain.PricingRequest, config map[string]interface{}) string {
	if currency, ok := domain.GetString(req.Inputs, "currency"); ok {
		return currency
	}
	if currency, ok := domain.GetString(config, "currency"); ok {
		return currency
	}
	return "USD"
}
