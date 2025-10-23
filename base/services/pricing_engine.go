package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/saintparish4/pantera/base/models"
)

type PricingEngine struct {
	// Add any dependencies here (e.g., database, cache)
}

func NewPricingEngine() *PricingEngine {
	return &PricingEngine{}
}

// CalculatePrice is the main entry point for price calculations
func (e *PricingEngine) CalculatePrice(rule models.PricingRule, req models.PriceRequest) (models.PriceResponse, error) {
	var price float64
	var err error
	var multiplier float64
	var metadata = make(map[string]interface{})
	var timeFactors *models.TimeFactors
	var gemstoneBreakdown *models.GemstoneBreakdown

	// Use provided timestamp or current time
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	switch rule.Strategy {
	case "cost_plus":
		price, err = e.costPlus(rule)
		multiplier = 1.0 + (rule.MarkupPercentage / 100.0)

	case "demand_based":
		price, multiplier, err = e.demandBased(rule, req.DemandLevel)
		metadata["demand_level"] = req.DemandLevel

	case "competitive":
		price, err = e.competitive(rule, req.CompetitorPrice)
		if req.CompetitorPrice > 0 {
			multiplier = 1.0 + (rule.MarkupPercentage / 100.0)
			metadata["competitor_price"] = req.CompetitorPrice
		}

	case "geographic":
		price, multiplier, err = e.geographic(rule, req.Location)
		metadata["location"] = req.Location
		metadata["base_region"] = e.getBaseRegion(req.Location)

	case "time_based":
		price, multiplier, timeFactors, err = e.timeBased(rule, timestamp, req.EventDate, req.CurrentDemand)
		metadata["timestamp"] = timestamp.Format(time.RFC3339)
		if req.EventDate != "" {
			metadata["event_date"] = req.EventDate
		}

	case "gemstone":
		price, multiplier, gemstoneBreakdown, err = e.gemstone(rule, req)
		metadata["gemstone_type"] = req.GemstoneType
		metadata["carat_weight"] = req.CaratWeight
		metadata["cut_grade"] = req.CutGrade
		metadata["clarity_grade"] = req.ClarityGrade

	default:
		return models.PriceResponse{}, errors.New("unsupported pricing strategy")
	}

	if err != nil {
		return models.PriceResponse{}, err
	}

	// Apply min/max bounds
	originalPrice := price
	price = e.applyBounds(price, rule)

	// Calculate discount if price was capped
	discount := 0.0
	if originalPrice > price {
		discount = originalPrice - price
	}

	// Get currency
	currency := rule.DefaultCurrency
	if currency == "" {
		currency = "USD"
	}

	// For gemstone pricing, use base price per carat as "original price"
	var displayOriginalPrice float64
	if gemstoneBreakdown != nil {
		displayOriginalPrice = gemstoneBreakdown.BasePricePerCarat * req.CaratWeight
	} else {
		displayOriginalPrice = rule.BasePrice
	}

	return models.PriceResponse{
		Price:             roundToTwoDecimals(price),
		OriginalPrice:     roundToTwoDecimals(displayOriginalPrice),
		Strategy:          rule.Strategy,
		Currency:          currency,
		Location:          req.Location,
		Discount:          roundToTwoDecimals(discount),
		Multiplier:        roundToTwoDecimals(multiplier),
		AppliedRule:       rule.Name,
		TimeFactors:       timeFactors,
		GemstoneBreakdown: gemstoneBreakdown,
		Metadata:          metadata,
	}, nil
}

// timeBased calculates price based on time factors
func (e *PricingEngine) timeBased(rule models.PricingRule, timestamp time.Time, eventDate string, currentDemand float64) (float64, float64, *models.TimeFactors, error) {
	timeRules := rule.TimeRules
	price := rule.BasePrice
	totalMultiplier := 1.0

	timeFactors := &models.TimeFactors{
		CalculatedAt:       timestamp.Format("2006-01-02 15:04:05"),
		AppliedMultipliers: []string{},
	}

	hour := timestamp.Hour()
	weekday := int(timestamp.Weekday())
	month := int(timestamp.Month())
	dateStr := timestamp.Format("2006-01-02")

	// 1. Check for special events (highest priority)
	for _, event := range timeRules.SpecialEvents {
		if event.Date == dateStr {
			// Check if current hour is within event time
			if e.isHourInRange(hour, event.StartHour, event.EndHour) {
				totalMultiplier *= event.Multiplier
				timeFactors.IsSpecialEvent = true
				timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers,
					"special_event:"+event.Name)
			}
		}
	}

	// 2. Check for holidays
	if e.contains(timeRules.HolidayDates, dateStr) && timeRules.HolidayMultiplier > 0 {
		totalMultiplier *= timeRules.HolidayMultiplier
		timeFactors.IsHoliday = true
		timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "holiday")
	}

	// 3. Check for peak season (month-based)
	if e.containsInt(timeRules.PeakMonths, month) && timeRules.PeakMonthMultiplier > 0 {
		totalMultiplier *= timeRules.PeakMonthMultiplier
		timeFactors.IsPeakSeason = true
		timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "peak_season")
	}

	// 4. Check for weekend
	if e.containsInt(timeRules.WeekendDays, weekday) && timeRules.WeekendMultiplier > 0 {
		totalMultiplier *= timeRules.WeekendMultiplier
		timeFactors.IsWeekend = true
		timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "weekend")
	}

	// 5. Check for peak hours
	if e.containsInt(timeRules.PeakHours, hour) && timeRules.PeakMultiplier > 0 {
		totalMultiplier *= timeRules.PeakMultiplier
		timeFactors.IsPeakHour = true
		timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "peak_hour")
	}

	// 6. Check for late night
	if e.containsInt(timeRules.LateNightHours, hour) && timeRules.LateNightMultiplier > 0 {
		totalMultiplier *= timeRules.LateNightMultiplier
		timeFactors.IsLateNight = true
		timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "late_night")
	}

	// 7. Event-based pricing (tickets, etc.)
	if eventDate != "" {
		eventMultiplier := e.calculateEventMultiplier(timestamp, eventDate, timeRules)
		if eventMultiplier > 1.0 {
			totalMultiplier *= eventMultiplier
			timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "event_timing")
		}
	}

	// 8. Apply surge pricing if enabled
	if rule.SurgeEnabled && currentDemand > 0 {
		surgeMultiplier := e.calculateSurge(currentDemand, rule.BaseSurgeThreshold)
		if surgeMultiplier > 1.0 {
			totalMultiplier *= surgeMultiplier
			timeFactors.SurgeActive = true
			timeFactors.SurgeLevel = surgeMultiplier
			timeFactors.AppliedMultipliers = append(timeFactors.AppliedMultipliers, "surge")
		}
	}

	// Calculate final price
	price = price * totalMultiplier

	return price, totalMultiplier, timeFactors, nil
}

// calculateEventMultiplier for tickets/events based on days before
func (e *PricingEngine) calculateEventMultiplier(now time.Time, eventDateStr string, rules models.TimeRules) float64 {
	eventDate, err := time.Parse("2006-01-02", eventDateStr)
	if err != nil {
		return 1.0
	}

	daysUntilEvent := int(eventDate.Sub(now).Hours() / 24)

	// Early bird discount
	if rules.EarlyBirdDaysBefore > 0 && daysUntilEvent >= rules.EarlyBirdDaysBefore {
		return rules.EarlyBirdMultiplier
	}

	// Last minute pricing
	if rules.LastMinuteDaysBefore > 0 && daysUntilEvent <= rules.LastMinuteDaysBefore && daysUntilEvent > 0 {
		return rules.LastMinuteMultiplier
	}

	// Day of event
	if daysUntilEvent == 0 && rules.DayOfEventMultiplier > 0 {
		return rules.DayOfEventMultiplier
	}

	return 1.0
}

// calculateSurge applies surge pricing based on current demand
func (e *PricingEngine) calculateSurge(currentDemand float64, threshold float64) float64 {
	if currentDemand <= threshold {
		return 1.0
	}

	// Simple linear surge: demand 1.5x threshold = 1.5x price
	// Could be made more sophisticated with exponential curves
	return currentDemand / threshold
}

// isHourInRange checks if hour is within range (handles midnight wraparound)
func (e *PricingEngine) isHourInRange(hour, start, end int) bool {
	if start <= end {
		return hour >= start && hour <= end
	}
	// Wraparound case (e.g., 22:00 to 02:00)
	return hour >= start || hour <= end
}

// costPlus applies a simple markup to the base price
func (e *PricingEngine) costPlus(rule models.PricingRule) (float64, error) {
	markup := rule.BasePrice * (rule.MarkupPercentage / 100.0)
	return rule.BasePrice + markup, nil
}

// demandBased adjusts price based on demand level
func (e *PricingEngine) demandBased(rule models.PricingRule, demandLevel float64) (float64, float64, error) {
	if demandLevel < 0 {
		demandLevel = 1.0 // Default to normal demand
	}

	// Use demand_multiplier from rule or default to 1.0
	baseMultiplier := rule.DemandMultiplier
	if baseMultiplier == 0 {
		baseMultiplier = 1.0
	}

	// Calculate final multiplier
	multiplier := baseMultiplier * demandLevel
	price := rule.BasePrice * multiplier

	return price, multiplier, nil
}

// competitive positions price against competitor pricing
func (e *PricingEngine) competitive(rule models.PricingRule, competitorPrice float64) (float64, error) {
	if competitorPrice <= 0 {
		return 0, errors.New("competitor_price must be provided and greater than 0")
	}

	// Apply markup/markdown percentage to competitor price
	adjustment := competitorPrice * (rule.MarkupPercentage / 100.0)
	price := competitorPrice + adjustment

	return price, nil
}

// geographic adjusts price based on customer location
func (e *PricingEngine) geographic(rule models.PricingRule, location string) (float64, float64, error) {
	if location == "" {
		return 0, 0, errors.New("location must be provided for geographic pricing")
	}

	if len(rule.RegionMultipliers) == 0 {
		return 0, 0, errors.New("region_multipliers not configured for this rule")
	}

	// Normalize location (uppercase)
	location = strings.ToUpper(strings.TrimSpace(location))

	// Try exact match first (e.g., "US-CA")
	multiplier, exists := rule.RegionMultipliers[location]

	if !exists {
		// Try country code only (e.g., "US" from "US-CA")
		countryCode := e.getBaseRegion(location)
		multiplier, exists = rule.RegionMultipliers[countryCode]

		if !exists {
			// Default to 1.0 if region not found
			multiplier = 1.0
		}
	}

	price := rule.BasePrice * multiplier
	return price, multiplier, nil
}

// getBaseRegion extracts the country code from a location string
func (e *PricingEngine) getBaseRegion(location string) string {
	parts := strings.Split(location, "-")
	if len(parts) > 0 {
		return strings.ToUpper(parts[0])
	}
	return strings.ToUpper(location)
}

// gemstone calculates price based on gem attributes
func (e *PricingEngine) gemstone(rule models.PricingRule, req models.PriceRequest) (float64, float64, *models.GemstoneBreakdown, error) {
	config := rule.GemstoneConfig

	// Validate required fields
	if req.GemstoneType == "" {
		return 0, 0, nil, errors.New("gemstone_type is required")
	}
	if req.CaratWeight <= 0 {
		return 0, 0, nil, errors.New("carat_weight must be greater than 0")
	}
	if req.CutGrade == "" {
		return 0, 0, nil, errors.New("cut_grade is required")
	}
	if req.ClarityGrade == "" {
		return 0, 0, nil, errors.New("clarity_grade is required")
	}

	// Normalize inputs
	gemType := strings.ToLower(strings.TrimSpace(req.GemstoneType))
	cutGrade := strings.ToLower(strings.TrimSpace(req.CutGrade))
	clarityGrade := strings.ToUpper(strings.TrimSpace(req.ClarityGrade))
	colorGrade := strings.ToUpper(strings.TrimSpace(req.ColorGrade))

	// Initialize breakdown
	breakdown := &models.GemstoneBreakdown{
		GemstoneType: gemType,
		CaratWeight:  req.CaratWeight,
		CutGrade:     cutGrade,
		ClarityGrade: clarityGrade,
		ColorGrade:   colorGrade,
	}

	// 1. Get base price per carat for gem type
	basePricePerCarat, exists := config.BasePricePerCarat[gemType]
	if !exists {
		return 0, 0, nil, fmt.Errorf("unsupported gemstone type: %s", gemType)
	}
	breakdown.BasePricePerCarat = basePricePerCarat

	// 2. Get carat tier multiplier
	caratMultiplier := e.getCaratTierMultiplier(req.CaratWeight, config.CaratTiers)
	breakdown.CaratTier = e.getCaratTierName(req.CaratWeight, config.CaratTiers)
	breakdown.CaratMultiplier = caratMultiplier

	// 3. Get cut quality multiplier
	cutMultiplier, exists := config.CutMultipliers[cutGrade]
	if !exists {
		return 0, 0, nil, fmt.Errorf("invalid cut grade: %s", cutGrade)
	}
	breakdown.CutMultiplier = cutMultiplier

	// 4. Get clarity multiplier
	clarityMultiplier, exists := config.ClarityMultipliers[clarityGrade]
	if !exists {
		return 0, 0, nil, fmt.Errorf("invalid clarity grade: %s", clarityGrade)
	}
	breakdown.ClarityMultiplier = clarityMultiplier

	// 5. Get color multiplier (optional for some gems)
	colorMultiplier := 1.0
	if colorGrade != "" && len(config.ColorMultipliers) > 0 {
		if mult, exists := config.ColorMultipliers[colorGrade]; exists {
			colorMultiplier = mult
			breakdown.ColorMultiplier = colorMultiplier
		} else {
			colorMultiplier = 1.0
			breakdown.ColorMultiplier = 1.0
		}
	} else {
		breakdown.ColorMultiplier = 1.0
	}

	// 6. Get certification multiplier (optional)
	certMultiplier := 1.0
	if req.Certification != "" && len(config.CertificationMultipliers) > 0 {
		cert := strings.ToUpper(req.Certification)
		if mult, exists := config.CertificationMultipliers[cert]; exists {
			certMultiplier = mult
			breakdown.Certification = cert
			breakdown.CertMultiplier = certMultiplier
		}
	}

	// 7. Get treatment multiplier (optional)
	treatmentMultiplier := 1.0
	if req.Treatment != "" && len(config.TreatmentMultipliers) > 0 {
		treatment := strings.ToLower(req.Treatment)
		if mult, exists := config.TreatmentMultipliers[treatment]; exists {
			treatmentMultiplier = mult
			breakdown.Treatment = treatment
			breakdown.TreatmentMultiplier = treatmentMultiplier
		}
	}

	// 8. Get origin multiplier (optional, for colored gems)
	originMultiplier := 1.0
	if req.Origin != "" && len(config.OriginMultipliers) > 0 {
		origin := strings.ToLower(req.Origin)
		if mult, exists := config.OriginMultipliers[origin]; exists {
			originMultiplier = mult
			breakdown.Origin = origin
			breakdown.OriginMultiplier = originMultiplier
		}
	}

	// Calculate total multiplier
	totalMultiplier := caratMultiplier * cutMultiplier * clarityMultiplier *
		colorMultiplier * certMultiplier * treatmentMultiplier * originMultiplier
	breakdown.TotalMultiplier = totalMultiplier

	// Calculate price per carat with all multipliers
	pricePerCarat := basePricePerCarat * totalMultiplier
	breakdown.PricePerCarat = pricePerCarat

	// Calculate total price
	totalPrice := pricePerCarat * req.CaratWeight

	return totalPrice, totalMultiplier, breakdown, nil
}

// getCaratTierMultiplier determines the multiplier based on carat weight
func (e *PricingEngine) getCaratTierMultiplier(carat float64, tiers map[string]float64) float64 {
	// Expected tier format: "0.5-1.0", "1.0-2.0", "2.0-3.0", "3.0+"

	if carat >= 3.0 {
		if mult, exists := tiers["3.0+"]; exists {
			return mult
		}
	}

	if carat >= 2.0 && carat < 3.0 {
		if mult, exists := tiers["2.0-3.0"]; exists {
			return mult
		}
	}

	if carat >= 1.0 && carat < 2.0 {
		if mult, exists := tiers["1.0-2.0"]; exists {
			return mult
		}
	}

	if carat >= 0.5 && carat < 1.0 {
		if mult, exists := tiers["0.5-1.0"]; exists {
			return mult
		}
	}

	// Default to 1.0 if no tier matches
	return 1.0
}

// getCaratTierName returns the tier name for display
func (e *PricingEngine) getCaratTierName(carat float64, tiers map[string]float64) string {
	if carat >= 3.0 {
		return "3.0+"
	}
	if carat >= 2.0 && carat < 3.0 {
		return "2.0-3.0"
	}
	if carat >= 1.0 && carat < 2.0 {
		return "1.0-2.0"
	}
	if carat >= 0.5 && carat < 1.0 {
		return "0.5-1.0"
	}
	return "under_0.5"
}

// applyBounds ensures price is within min/max constraints
func (e *PricingEngine) applyBounds(price float64, rule models.PricingRule) float64 {
	if rule.MinPrice > 0 && price < rule.MinPrice {
		return rule.MinPrice
	}
	if rule.MaxPrice > 0 && price > rule.MaxPrice {
		return rule.MaxPrice
	}
	return price
}

// roundToTwoDecimals rounds a float to 2 decimal places
func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100.0
}

// contains checks if a string slice contains a value
func (e *PricingEngine) contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// containsInt checks if an int slice contains a value
func (e *PricingEngine) containsInt(arr []int, val int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// LogCalculation creates an audit record of the price calculation
func (e *PricingEngine) LogCalculation(rule models.PricingRule, req models.PriceRequest, response models.PriceResponse) models.PriceCalculation {
	inputDataBytes, _ := json.Marshal(req)

	location := &req.Location
	if req.Location == "" {
		location = nil
	}

	currency := &response.Currency

	// Extract time information
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	hour := timestamp.Hour()
	dayOfWeek := int(timestamp.Weekday())

	var surgeMultiplier *float64
	if response.TimeFactors != nil && response.TimeFactors.SurgeActive {
		surgeMultiplier = &response.TimeFactors.SurgeLevel
	}

	// Gemstone fields
	var gemstoneType *string
	var caratWeight *float64
	var cutGrade *string
	var clarityGrade *string
	var colorGrade *string
	var certification *string

	if req.GemstoneType != "" {
		gemstoneType = &req.GemstoneType
		caratWeight = &req.CaratWeight
		cutGrade = &req.CutGrade
		clarityGrade = &req.ClarityGrade
		if req.ColorGrade != "" {
			colorGrade = &req.ColorGrade
		}
		if req.Certification != "" {
			certification = &req.Certification
		}
	}

	return models.PriceCalculation{
		RuleID:                rule.ID,
		InputData:             string(inputDataBytes),
		CalculatedPrice:       response.Price,
		StrategyUsed:          rule.Strategy,
		Location:              location,
		Currency:              currency,
		CalculatedAtHour:      &hour,
		CalculatedAtDayOfWeek: &dayOfWeek,
		SurgeMultiplier:       surgeMultiplier,
		GemstoneType:          gemstoneType,
		CaratWeight:           caratWeight,
		CutGrade:              cutGrade,
		ClarityGrade:          clarityGrade,
		ColorGrade:            colorGrade,
		Certification:         certification,
		CreatedAt:             timestamp,
	}
}
