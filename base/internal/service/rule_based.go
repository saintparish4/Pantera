package service

import (
	"fmt"
	"strings"

	"github.com/saintparish4/harmonia/internal/domain"
)

// RuleBasedStrategy implements custom rule-based pricing logic
type RuleBasedStrategy struct{}

// Name returns the strategy identifier
func (s *RuleBasedStrategy) Name() string {
	return domain.StrategyTypeRuleBased
}

// PricingRule represents a single pricing rule
type PricingRule struct {
	Condition string                 `json:"condition"` // e.g., "quantity > 10"
	Action    string                 `json:"action"`    // e.g., "apply_discount", "set_multiplier"
	Value     float64                `json:"value"`     // The value to apply
	Priority  int                    `json:"priority"`  // Higher priority rules execute first
}

// Validate checks if the configuration is valid for rule-based pricing
func (s *RuleBasedStrategy) Validate(config map[string]interface{}) error {
	// rules array is required in config
	rulesData, ok := config["rules"]
	if !ok {
		return fmt.Errorf("%w: rules array is required", domain.ErrConfigurationInvalid)
	}
	
	// Convert to slice
	rules, ok := rulesData.([]interface{})
	if !ok {
		return fmt.Errorf("rules must be an array")
	}
	
	if len(rules) == 0 {
		return fmt.Errorf("rules array cannot be empty")
	}
	
	// Validate each rule
	for i, rule := range rules {
		ruleMap, ok := rule.(map[string]interface{})
		if !ok {
			return fmt.Errorf("rules[%d] must be an object", i)
		}
		
		// Validate required fields
		if _, ok := domain.GetString(ruleMap, "condition"); !ok {
			return fmt.Errorf("rules[%d] missing required field: condition", i)
		}
		
		if _, ok := domain.GetString(ruleMap, "action"); !ok {
			return fmt.Errorf("rules[%d] missing required field: action", i)
		}
		
		if _, ok := domain.GetFloat64(ruleMap, "value"); !ok {
			return fmt.Errorf("rules[%d] missing required field: value", i)
		}
		
		// Validate action type
		action, _ := domain.GetString(ruleMap, "action")
		validActions := map[string]bool{
			"apply_discount":   true,
			"apply_markup":     true,
			"set_multiplier":   true,
			"add_fixed_amount": true,
			"set_price":        true,
		}
		
		if !validActions[action] {
			return fmt.Errorf("rules[%d] has invalid action: %s", i, action)
		}
	}
	
	return nil
}

// Calculate computes the rule-based price
func (s *RuleBasedStrategy) Calculate(req *domain.PricingRequest, config map[string]interface{}) (*domain.PricingResponse, error) {
	// Extract base_price from inputs (required)
	basePrice, ok := domain.GetFloat64(req.Inputs, "base_price")
	if !ok {
		return nil, fmt.Errorf("%w: base_price is required", domain.ErrMissingRequiredField)
	}
	
	if basePrice < 0 {
		return nil, fmt.Errorf("%w: base_price cannot be negative", domain.ErrInvalidFieldValue)
	}
	
	// Get context for rule evaluation
	context := req.Inputs
	
	// Get rules from config
	rulesData := config["rules"]
	rules, _ := rulesData.([]interface{})
	
	// Evaluate rules and build adjustments
	currentPrice := basePrice
	adjustments := []domain.PriceAdjustment{}
	rulesApplied := []map[string]interface{}{}
	
	for _, ruleData := range rules {
		ruleMap, ok := ruleData.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Parse rule
		rule := PricingRule{}
		rule.Condition, _ = domain.GetString(ruleMap, "condition")
		rule.Action, _ = domain.GetString(ruleMap, "action")
		rule.Value, _ = domain.GetFloat64(ruleMap, "value")
		rule.Priority, _ = ruleMap["priority"].(int)
		
		// Evaluate condition
		if !s.evaluateCondition(rule.Condition, context) {
			continue
		}
		
		// Apply action
		priceBeforeAction := currentPrice
		currentPrice = s.applyAction(currentPrice, rule)
		adjustment := currentPrice - priceBeforeAction
		
		// Record adjustment
		adjustments = append(adjustments, domain.PriceAdjustment{
			Type:        rule.Action,
			Description: fmt.Sprintf("Rule: %s", rule.Condition),
			Amount:      rule.Value,
			Applied:     adjustment,
		})
		
		// Track applied rule
		rulesApplied = append(rulesApplied, map[string]interface{}{
			"condition": rule.Condition,
			"action":    rule.Action,
			"value":     rule.Value,
			"result":    domain.RoundToTwoDecimals(currentPrice),
		})
	}
	
	finalPrice := currentPrice
	
	// Apply min/max bounds
	minPrice, _ := domain.GetFloat64(config, "min_price")
	maxPrice, _ := domain.GetFloat64(config, "max_price")
	finalPrice = domain.ApplyBounds(finalPrice, minPrice, maxPrice)
	
	// Build response
	response := &domain.PricingResponse{
		FinalPrice:    finalPrice,
		OriginalPrice: basePrice,
		Currency:      getCurrency(req, config),
		Breakdown: domain.PriceBreakdown{
			BasePrice:   basePrice,
			Adjustments: adjustments,
			Details: map[string]interface{}{
				"base_price":    basePrice,
				"rules_applied": rulesApplied,
				"final_price":   domain.RoundToTwoDecimals(finalPrice),
			},
		},
	}
	
	return response, nil
}

// evaluateCondition evaluates a condition string against the context
// Supports simple comparisons: quantity > 10, customer_tier == "premium", total_value >= 100
func (s *RuleBasedStrategy) evaluateCondition(condition string, context map[string]interface{}) bool {
	condition = strings.TrimSpace(condition)
	
	// Always true condition
	if condition == "true" || condition == "always" {
		return true
	}
	
	// Parse condition: "field operator value"
	// Supported operators: >, <, >=, <=, ==, !=
	
	// Try >= first (before >)
	if parts := strings.Split(condition, ">="); len(parts) == 2 {
		return s.compareValues(parts[0], parts[1], context, ">=")
	}
	
	// Try <= (before <)
	if parts := strings.Split(condition, "<="); len(parts) == 2 {
		return s.compareValues(parts[0], parts[1], context, "<=")
	}
	
	// Try == (before single =)
	if parts := strings.Split(condition, "=="); len(parts) == 2 {
		return s.compareValues(parts[0], parts[1], context, "==")
	}
	
	// Try !=
	if parts := strings.Split(condition, "!="); len(parts) == 2 {
		return s.compareValues(parts[0], parts[1], context, "!=")
	}
	
	// Try >
	if parts := strings.Split(condition, ">"); len(parts) == 2 {
		return s.compareValues(parts[0], parts[1], context, ">")
	}
	
	// Try <
	if parts := strings.Split(condition, "<"); len(parts) == 2 {
		return s.compareValues(parts[0], parts[1], context, "<")
	}
	
	// Unknown condition format, return false
	return false
}

// compareValues compares a field from context with a target value
func (s *RuleBasedStrategy) compareValues(fieldName, targetValue string, context map[string]interface{}, operator string) bool {
	fieldName = strings.TrimSpace(fieldName)
	targetValue = strings.TrimSpace(targetValue)
	
	// Get field value from context
	fieldValue, exists := context[fieldName]
	if !exists {
		return false
	}
	
	// Try numeric comparison
	if fieldNum, ok := convertToFloat(fieldValue); ok {
		// Try to parse target as number
		if targetNum, ok := parseFloat(targetValue); ok {
			switch operator {
			case ">":
				return fieldNum > targetNum
			case "<":
				return fieldNum < targetNum
			case ">=":
				return fieldNum >= targetNum
			case "<=":
				return fieldNum <= targetNum
			case "==":
				return fieldNum == targetNum
			case "!=":
				return fieldNum != targetNum
			}
		}
	}
	
	// String comparison
	fieldStr := fmt.Sprintf("%v", fieldValue)
	targetStr := strings.Trim(targetValue, "\"'")
	
	switch operator {
	case "==":
		return fieldStr == targetStr
	case "!=":
		return fieldStr != targetStr
	}
	
	return false
}

// applyAction applies a pricing action to the current price
func (s *RuleBasedStrategy) applyAction(currentPrice float64, rule PricingRule) float64 {
	switch rule.Action {
	case "apply_discount":
		// Value is percentage (e.g., 15 for 15% off)
		discount := currentPrice * (rule.Value / 100)
		return currentPrice - discount
		
	case "apply_markup":
		// Value is percentage (e.g., 20 for 20% markup)
		markup := currentPrice * (rule.Value / 100)
		return currentPrice + markup
		
	case "set_multiplier":
		// Value is multiplier (e.g., 1.5 for 1.5x)
		return currentPrice * rule.Value
		
	case "add_fixed_amount":
		// Value is fixed amount to add (can be negative)
		return currentPrice + rule.Value
		
	case "set_price":
		// Value is the new price
		return rule.Value
	}
	
	return currentPrice
}

// parseFloat parses a string to float64
func parseFloat(s string) (float64, bool) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err == nil
}