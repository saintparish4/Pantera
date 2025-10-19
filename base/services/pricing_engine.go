package services

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/saintparish4/pantera/base/database"
	"github.com/saintparish4/pantera/base/models"
)

type PricingEngine struct{}

func NewPricingEngine() *PricingEngine {
	return &PricingEngine{}
}

// CalculatePrice - Main entry point for pricing calculations
func (pe *PricingEngine) CalculatePrice(req models.PriceRequest) (*models.PriceResponse, error) {
	// Get pricing rule from database
	rule, err := pe.GetRule(req.RuleID)
	if err != nil {
		return nil, err
	}

	if !rule.Active {
		return nil, fmt.Errorf("pricing rule is not active")
	}

	// Calculate based on strategy
	var price float64
	var breakdown map[string]interface{}

	switch rule.Strategy {
	case "cost_plus":
		price, breakdown = pe.CostPlusStrategy(*rule, req)
	case "demand_based":
		price, breakdown = pe.DemandBasedStrategy(*rule, req)
	case "competitive":
		price, breakdown = pe.CompetitiveStrategy(*rule, req)
	default:
		return nil, fmt.Errorf("invalid pricing strategy: %s", rule.Strategy)
	}

	// Apply min/max constraints
	originalPrice := price
	price = pe.ApplyConstraints(price, rule.MinPrice, rule.MaxPrice)

	// Log calculation for audit
	pe.LogCalculation(rule.ID, req, price, rule.Strategy)

	return &models.PriceResponse{
		Price:         roundToTwoDecimals(price),
		OriginalPrice: roundToTwoDecimals(originalPrice),
		Strategy:      rule.Strategy,
		Breakdown:     breakdown,
		CalculatedAt:  time.Now(),
	}, nil
}

// CostPlusStrategy - Base price + markup percentage
func (pe *PricingEngine) CostPlusStrategy(rule models.PricingRule, req models.PriceRequest) (float64, map[string]interface{}) {
	markup := rule.BasePrice * (rule.MarkupPercentage / 100)
	price := rule.BasePrice + markup

	breakdown := map[string]interface{}{
		"base_price":        rule.BasePrice,
		"markup_percentage": rule.MarkupPercentage,
		"markup_amount":     roundToTwoDecimals(markup),
	}

	return price, breakdown
}

// DemandBasedStrategy - Adjust price based on demand level
func (pe *PricingEngine) DemandBasedStrategy(rule models.PricingRule, req models.PriceRequest) (float64, map[string]interface{}) {
	// Start with cost-plus
	basePrice, _ := pe.CostPlusStrategy(rule, req)

	// Apply demand multiplier (0.0 = low demand, 1.0 = normal, 2.0 = high demand)
	demandLevel := float64(req.DemandLevel)
	if demandLevel == 0 {
		demandLevel = 1.0 // default to normal demand
	}

	// Use rule's demand multiplier as sensitivity factor
	sensitivity := rule.DemandMultiplier
	if sensitivity == 0 {
		sensitivity = 1.0
	}

	demandAdjustment := (demandLevel - 1.0) * sensitivity
	price := basePrice * (1 + demandAdjustment)

	breakdown := map[string]interface{}{
		"base_price":         roundToTwoDecimals(basePrice),
		"demand_level":       demandLevel,
		"demand_adjustment":  fmt.Sprintf("%.2f%%", demandAdjustment*100),
		"sensitivity_factor": sensitivity,
	}

	return price, breakdown
}

// CompetitiveStrategy - Price based on competitor pricing
func (pe *PricingEngine) CompetitiveStrategy(rule models.PricingRule, req models.PriceRequest) (float64, map[string]interface{}) {
	if req.CompetitorPrice == 0 {
		// Fallback to cost-plus if no competitor price provided
		return pe.CostPlusStrategy(rule, req)
	}

	// Use markup percentage as competitive positioning
	// Negative = undercut, Positive = premium
	competitiveAdjustment := req.CompetitorPrice * (rule.MarkupPercentage / 100)
	price := req.CompetitorPrice + competitiveAdjustment

	breakdown := map[string]interface{}{
		"competitor_price":       req.CompetitorPrice,
		"positioning_percentage": rule.MarkupPercentage,
		"adjustment_amount":      roundToTwoDecimals(competitiveAdjustment),
		"positioning":            getPositioningLabel(rule.MarkupPercentage),
	}

	return price, breakdown
}

// ApplyConstraints - Ensure price stays within min/max bounds
func (pe *PricingEngine) ApplyConstraints(price, minPrice, maxPrice float64) float64 {
	if minPrice > 0 && price < minPrice {
		return minPrice
	}
	if maxPrice > 0 && price > maxPrice {
		return maxPrice
	}
	return price
}

// GetRule - Fetch pricing rule from database
func (pe *PricingEngine) GetRule(ruleID int) (*models.PricingRule, error) {
	query := `
		SELECT id, name, strategy, base_price, markup_percentage, 
		       min_price, max_price, demand_multiplier, active, created_at, updated_at
		FROM pricing_rules
		WHERE id = $1
	`

	var rule models.PricingRule
	err := database.DB.QueryRow(query, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Strategy, &rule.BasePrice,
		&rule.MarkupPercentage, &rule.MinPrice, &rule.MaxPrice,
		&rule.DemandMultiplier, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("rule not found: %v", err)
	}

	return &rule, nil
}

// LogCalculation - Save calculation to audit log
func (pe *PricingEngine) LogCalculation(ruleID int, req models.PriceRequest, price float64, strategy string) error {
	inputData, _ := json.Marshal(req)

	query := `
		INSERT INTO price_calculations (rule_id, input_data, calculated_price, strategy_used)
		VALUES ($1, $2, $3, $4)
	`

	_, err := database.DB.Exec(query, ruleID, inputData, price, strategy)
	return err
}

// Helper functions
func roundToTwoDecimals(num float64) float64 {
	return math.Round(num*100) / 100
}

func getPositioningLabel(percentage float64) string {
	if percentage < -5 {
		return "aggressive_undercut"
	} else if percentage < 0 {
		return "slight_undercut"
	} else if percentage < 5 {
		return "competitive"
	} else if percentage < 15 {
		return "premium"
	}
	return "luxury"
}
