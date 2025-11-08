package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/saintparish4/harmonia/internal/domain"
)

// TimeBasedStrategy implements time-of-day, day-of-week, and surge pricing
type TimeBasedStrategy struct{}

// Name returns the strategy identifier
func (s *TimeBasedStrategy) Name() string {
	return domain.StrategyTypeTimeBased
}

// TimeWindow represents a time-based pricing rule
type TimeWindow struct {
	Days       []string  `json:"days"`        // e.g., ["monday", "friday"]
	StartTime  string    `json:"start_time"`  // e.g., "17:00"
	EndTime    string    `json:"end_time"`    // e.g., "20:00"
	Multiplier float64   `json:"multiplier"`  // e.g., 1.5
}

// Validate checks if the configuration is valid for time-based pricing
func (s *TimeBasedStrategy) Validate(config map[string]interface{}) error {
	// time_windows is required in config
	timeWindowsData, ok := config["time_windows"]
	if !ok {
		return fmt.Errorf("%w: time_windows array is required", domain.ErrConfigurationInvalid)
	}
	
	// Convert to slice
	windows, ok := timeWindowsData.([]interface{})
	if !ok {
		return fmt.Errorf("time_windows must be an array")
	}
	
	if len(windows) == 0 {
		return fmt.Errorf("time_windows array cannot be empty")
	}
	
	// Validate each time window
	for i, window := range windows {
		windowMap, ok := window.(map[string]interface{})
		if !ok {
			return fmt.Errorf("time_windows[%d] must be an object", i)
		}
		
		// Validate required fields
		if _, ok := windowMap["multiplier"]; !ok {
			return fmt.Errorf("time_windows[%d] missing required field: multiplier", i)
		}
		
		// Validate time format if provided
		if startTime, ok := domain.GetString(windowMap, "start_time"); ok {
			if !isValidTimeFormat(startTime) {
				return fmt.Errorf("time_windows[%d] has invalid start_time format (use HH:MM)", i)
			}
		}
		
		if endTime, ok := domain.GetString(windowMap, "end_time"); ok {
			if !isValidTimeFormat(endTime) {
				return fmt.Errorf("time_windows[%d] has invalid end_time format (use HH:MM)", i)
			}
		}
	}
	
	return nil
}

// Calculate computes the time-based price
func (s *TimeBasedStrategy) Calculate(req *domain.PricingRequest, config map[string]interface{}) (*domain.PricingResponse, error) {
	// Extract base_price from inputs (required)
	basePrice, ok := domain.GetFloat64(req.Inputs, "base_price")
	if !ok {
		return nil, fmt.Errorf("%w: base_price is required", domain.ErrMissingRequiredField)
	}
	
	if basePrice < 0 {
		return nil, fmt.Errorf("%w: base_price cannot be negative", domain.ErrInvalidFieldValue)
	}
	
	// Get timestamp (default to now)
	timestamp := req.RequestedAt
	if timestampStr, ok := domain.GetString(req.Inputs, "timestamp"); ok {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			timestamp = t
		}
	}
	
	// Get time windows from config
	timeWindowsData := config["time_windows"]
	windows, _ := timeWindowsData.([]interface{})
	
	// Find matching time windows and calculate multiplier
	activeWindows, totalMultiplier := s.findActiveWindows(timestamp, windows)
	
	// Check for surge pricing
	surgeMultiplier := 1.0
	currentDemand, hasDemand := domain.GetFloat64(req.Inputs, "current_demand")
	if hasDemand {
		surgeEnabled, _ := config["surge_enabled"].(bool)
		if surgeEnabled {
			surgeThreshold, _ := domain.GetFloat64(config, "base_surge_threshold")
			if surgeThreshold == 0 {
				surgeThreshold = 1.0
			}
			surgeMultiplier = calculateSurgeMultiplier(currentDemand, surgeThreshold)
		}
	}
	
	// Combine multipliers
	combinedMultiplier := totalMultiplier * surgeMultiplier
	finalPrice := basePrice * combinedMultiplier
	
	// Apply min/max bounds
	minPrice, _ := domain.GetFloat64(config, "min_price")
	maxPrice, _ := domain.GetFloat64(config, "max_price")
	finalPrice = domain.ApplyBounds(finalPrice, minPrice, maxPrice)
	
	// Build adjustments list
	adjustments := []domain.PriceAdjustment{}
	
	// Add time window adjustments
	for _, window := range activeWindows {
		adjustments = append(adjustments, domain.PriceAdjustment{
			Type:        "time_window",
			Description: fmt.Sprintf("%s surge (%s-%s)", strings.Join(window.Days, "/"), window.StartTime, window.EndTime),
			Amount:      window.Multiplier,
			Applied:     basePrice * (window.Multiplier - 1),
		})
	}
	
	// Add surge adjustment if applicable
	if surgeMultiplier > 1.0 {
		adjustments = append(adjustments, domain.PriceAdjustment{
			Type:        "surge",
			Description: fmt.Sprintf("Demand surge (%.1fx)", surgeMultiplier),
			Amount:      surgeMultiplier,
			Applied:     basePrice * totalMultiplier * (surgeMultiplier - 1),
		})
	}
	
	// Build response
	response := &domain.PricingResponse{
		FinalPrice:    finalPrice,
		OriginalPrice: basePrice,
		Currency:      getCurrency(req, config),
		Breakdown: domain.PriceBreakdown{
			BasePrice:   basePrice,
			Adjustments: adjustments,
			Details: map[string]interface{}{
				"base_price":           basePrice,
				"timestamp":            timestamp.Format(time.RFC3339),
				"day_of_week":          timestamp.Weekday().String(),
				"hour":                 timestamp.Hour(),
				"active_windows":       len(activeWindows),
				"time_multiplier":      totalMultiplier,
				"surge_multiplier":     surgeMultiplier,
				"combined_multiplier":  combinedMultiplier,
				"final_price":          domain.RoundToTwoDecimals(finalPrice),
			},
		},
	}
	
	if hasDemand {
		response.Breakdown.Details["current_demand"] = currentDemand
	}
	
	return response, nil
}

// findActiveWindows identifies which time windows apply to the given timestamp
func (s *TimeBasedStrategy) findActiveWindows(timestamp time.Time, windows []interface{}) ([]TimeWindow, float64) {
	activeWindows := []TimeWindow{}
	multiplier := 1.0
	
	dayName := strings.ToLower(timestamp.Weekday().String())
	currentTime := fmt.Sprintf("%02d:%02d", timestamp.Hour(), timestamp.Minute())
	
	for _, window := range windows {
		windowMap, ok := window.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Parse time window
		tw := TimeWindow{
			Multiplier: 1.0,
		}
		
		// Get multiplier
		if mult, ok := domain.GetFloat64(windowMap, "multiplier"); ok {
			tw.Multiplier = mult
		}
		
		// Get days
		if daysData, ok := windowMap["days"]; ok {
			if daysList, ok := daysData.([]interface{}); ok {
				for _, day := range daysList {
					if dayStr, ok := day.(string); ok {
						tw.Days = append(tw.Days, strings.ToLower(dayStr))
					}
				}
			}
		}
		
		// Get times
		tw.StartTime, _ = domain.GetString(windowMap, "start_time")
		tw.EndTime, _ = domain.GetString(windowMap, "end_time")
		
		// Check if window applies
		if s.windowApplies(dayName, currentTime, tw) {
			activeWindows = append(activeWindows, tw)
			multiplier *= tw.Multiplier
		}
	}
	
	return activeWindows, multiplier
}

// windowApplies checks if a time window applies to the given day and time
func (s *TimeBasedStrategy) windowApplies(dayName, currentTime string, window TimeWindow) bool {
	// Check day match
	dayMatches := len(window.Days) == 0 // If no days specified, applies to all days
	for _, day := range window.Days {
		if strings.ToLower(day) == dayName {
			dayMatches = true
			break
		}
	}
	
	if !dayMatches {
		return false
	}
	
	// Check time match
	if window.StartTime == "" || window.EndTime == "" {
		return true // If no time specified, applies all day
	}
	
	return currentTime >= window.StartTime && currentTime <= window.EndTime
}

// calculateSurgeMultiplier calculates surge pricing based on demand
func calculateSurgeMultiplier(currentDemand, baseThreshold float64) float64 {
	if currentDemand <= baseThreshold {
		return 1.0
	}
	
	// Linear surge: every 0.5 above threshold adds 0.25x
	// e.g., demand=1.5, threshold=1.0 -> multiplier=1.25
	excess := currentDemand - baseThreshold
	surgeMultiplier := 1.0 + (excess * 0.5)
	
	// Cap at 3x surge
	if surgeMultiplier > 3.0 {
		surgeMultiplier = 3.0
	}
	
	return surgeMultiplier
}

// isValidTimeFormat checks if a time string is in HH:MM format
func isValidTimeFormat(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}
	
	_, err := time.Parse("15:04", timeStr)
	return err == nil
}