package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// GemstoneConfig defines gemstone pricing configuration
type GemstoneConfig struct {
	BasePricePerCarat        map[string]float64 `json:"base_price_per_carat"`
	CaratTiers               map[string]float64 `json:"carat_tiers"`
	CutMultipliers           map[string]float64 `json:"cut_multipliers"`
	ClarityMultipliers       map[string]float64 `json:"clarity_multipliers"`
	ColorMultipliers         map[string]float64 `json:"color_multipliers"`
	CertificationMultipliers map[string]float64 `json:"certification_multipliers,omitempty"`
	TreatmentMultipliers     map[string]float64 `json:"treatment_multipliers,omitempty"`
	OriginMultipliers        map[string]float64 `json:"origin_multipliers,omitempty"`
}

// Value implements driver.Valuer for database storage
func (gc GemstoneConfig) Value() (driver.Value, error) {
	return json.Marshal(gc)
}

// Scan implements sql.Scanner for database retrieval
func (gc *GemstoneConfig) Scan(value interface{}) error {
	if value == nil {
		*gc = GemstoneConfig{
			BasePricePerCarat:        make(map[string]float64),
			CaratTiers:               make(map[string]float64),
			CutMultipliers:           make(map[string]float64),
			ClarityMultipliers:       make(map[string]float64),
			ColorMultipliers:         make(map[string]float64),
			CertificationMultipliers: make(map[string]float64),
			TreatmentMultipliers:     make(map[string]float64),
			OriginMultipliers:        make(map[string]float64),
		}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal GemstoneConfig")
	}

	return json.Unmarshal(bytes, gc)
}

// TimeRules defines time-based pricing configuration
type TimeRules struct {
	PeakHours            []int          `json:"peak_hours,omitempty"`
	PeakMultiplier       float64        `json:"peak_multiplier,omitempty"`
	LateNightHours       []int          `json:"late_night_hours,omitempty"`
	LateNightMultiplier  float64        `json:"late_night_multiplier,omitempty"`
	WeekendDays          []int          `json:"weekend_days,omitempty"`
	WeekendMultiplier    float64        `json:"weekend_multiplier,omitempty"`
	PeakMonths           []int          `json:"peak_months,omitempty"`
	PeakMonthMultiplier  float64        `json:"peak_month_multiplier,omitempty"`
	HolidayDates         []string       `json:"holiday_dates,omitempty"`
	HolidayMultiplier    float64        `json:"holiday_multiplier,omitempty"`
	SpecialEvents        []SpecialEvent `json:"special_events,omitempty"`
	EarlyBirdDaysBefore  int            `json:"early_bird_days_before,omitempty"`
	EarlyBirdMultiplier  float64        `json:"early_bird_multiplier,omitempty"`
	LastMinuteDaysBefore int            `json:"last_minute_days_before,omitempty"`
	LastMinuteMultiplier float64        `json:"last_minute_multiplier,omitempty"`
	DayOfEventMultiplier float64        `json:"day_of_event_multiplier,omitempty"`
}

// SpecialEvent defines a specific date/time event with custom pricing
type SpecialEvent struct {
	Name       string  `json:"name"`
	Date       string  `json:"date"`
	StartHour  int     `json:"start_hour"`
	EndHour    int     `json:"end_hour"`
	Multiplier float64 `json:"multiplier"`
}

// Value implements driver.Valuer for database storage
func (tr TimeRules) Value() (driver.Value, error) {
	return json.Marshal(tr)
}

// Scan implements sql.Scanner for database retrieval
func (tr *TimeRules) Scan(value interface{}) error {
	if value == nil {
		*tr = TimeRules{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal TimeRules")
	}

	return json.Unmarshal(bytes, tr)
}

// RegionMultipliers is a custom type for JSONB storage
type RegionMultipliers map[string]float64

// Value implements the driver.Valuer interface for database storage
func (rm RegionMultipliers) Value() (driver.Value, error) {
	if rm == nil {
		return json.Marshal(map[string]float64{})
	}
	return json.Marshal(rm)
}

// Scan implements the sql.Scanner interface for database retrieval
func (rm *RegionMultipliers) Scan(value interface{}) error {
	if value == nil {
		*rm = make(map[string]float64)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal RegionMultipliers")
	}

	return json.Unmarshal(bytes, rm)
}

// PricingRule represents a pricing configuration with time-based and gemstone support
type PricingRule struct {
	ID                 int               `json:"id"`
	Name               string            `json:"name"`
	Strategy           string            `json:"strategy"`
	BasePrice          float64           `json:"base_price"`
	MarkupPercentage   float64           `json:"markup_percentage"`
	MinPrice           float64           `json:"min_price"`
	MaxPrice           float64           `json:"max_price"`
	DemandMultiplier   float64           `json:"demand_multiplier,omitempty"`
	RegionMultipliers  RegionMultipliers `json:"region_multipliers,omitempty"`
	DefaultCurrency    string            `json:"default_currency,omitempty"`
	TimeRules          TimeRules         `json:"time_rules,omitempty"`
	SurgeEnabled       bool              `json:"surge_enabled,omitempty"`
	BaseSurgeThreshold float64           `json:"base_surge_threshold,omitempty"`
	GemstoneConfig     GemstoneConfig    `json:"gemstone_config,omitempty"`
	Active             bool              `json:"active"`
	CreatedAt          time.Time         `json:"created_at"`
}

// PriceCalculation represents a price calculation record with time tracking and gemstone support
type PriceCalculation struct {
	ID                    int       `json:"id"`
	RuleID                int       `json:"rule_id"`
	InputData             string    `json:"input_data"`
	CalculatedPrice       float64   `json:"calculated_price"`
	StrategyUsed          string    `json:"strategy_used"`
	Location              *string   `json:"location,omitempty"`
	Currency              *string   `json:"currency,omitempty"`
	CalculatedAtHour      *int      `json:"calculated_at_hour,omitempty"`
	CalculatedAtDayOfWeek *int      `json:"calculated_at_day_of_week,omitempty"`
	SurgeMultiplier       *float64  `json:"surge_multiplier,omitempty"`
	GemstoneType          *string   `json:"gemstone_type,omitempty"`
	CaratWeight           *float64  `json:"carat_weight,omitempty"`
	CutGrade              *string   `json:"cut_grade,omitempty"`
	ClarityGrade          *string   `json:"clarity_grade,omitempty"`
	ColorGrade            *string   `json:"color_grade,omitempty"`
	Certification         *string   `json:"certification,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

// PriceRequest represents an incoming price calculation request with time-based and gemstone parameters
type PriceRequest struct {
	RuleID          int       `json:"rule_id" binding:"required"`
	Quantity        int       `json:"quantity,omitempty"`
	DemandLevel     float64   `json:"demand_level,omitempty"`
	CompetitorPrice float64   `json:"competitor_price,omitempty"`
	Location        string    `json:"location,omitempty"`
	CustomerSegment string    `json:"customer_segment,omitempty"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
	EventDate       string    `json:"event_date,omitempty"`
	CurrentDemand   float64   `json:"current_demand,omitempty"`

	// Gemstone-specific fields
	GemstoneType  string  `json:"gemstone_type,omitempty"`
	CaratWeight   float64 `json:"carat_weight,omitempty"`
	CutGrade      string  `json:"cut_grade,omitempty"`
	ClarityGrade  string  `json:"clarity_grade,omitempty"`
	ColorGrade    string  `json:"color_grade,omitempty"`
	Certification string  `json:"certification,omitempty"`
	Treatment     string  `json:"treatment,omitempty"`
	Origin        string  `json:"origin,omitempty"`
}

// PriceResponse represents the API response with time-based and gemstone metadata
type PriceResponse struct {
	Price             float64                `json:"price"`
	OriginalPrice     float64                `json:"original_price,omitempty"`
	Strategy          string                 `json:"strategy"`
	Currency          string                 `json:"currency"`
	Location          string                 `json:"location,omitempty"`
	Discount          float64                `json:"discount,omitempty"`
	Multiplier        float64                `json:"multiplier,omitempty"`
	AppliedRule       string                 `json:"applied_rule,omitempty"`
	TimeFactors       *TimeFactors           `json:"time_factors,omitempty"`
	GemstoneBreakdown *GemstoneBreakdown     `json:"gemstone_breakdown,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// GemstoneBreakdown shows pricing calculation details for gemstones
type GemstoneBreakdown struct {
	GemstoneType        string  `json:"gemstone_type"`
	CaratWeight         float64 `json:"carat_weight"`
	BasePricePerCarat   float64 `json:"base_price_per_carat"`
	CaratTier           string  `json:"carat_tier"`
	CaratMultiplier     float64 `json:"carat_multiplier"`
	CutGrade            string  `json:"cut_grade"`
	CutMultiplier       float64 `json:"cut_multiplier"`
	ClarityGrade        string  `json:"clarity_grade"`
	ClarityMultiplier   float64 `json:"clarity_multiplier"`
	ColorGrade          string  `json:"color_grade"`
	ColorMultiplier     float64 `json:"color_multiplier"`
	Certification       string  `json:"certification,omitempty"`
	CertMultiplier      float64 `json:"cert_multiplier,omitempty"`
	Treatment           string  `json:"treatment,omitempty"`
	TreatmentMultiplier float64 `json:"treatment_multiplier,omitempty"`
	Origin              string  `json:"origin,omitempty"`
	OriginMultiplier    float64 `json:"origin_multiplier,omitempty"`
	TotalMultiplier     float64 `json:"total_multiplier"`
	PricePerCarat       float64 `json:"price_per_carat"`
}

// TimeFactors shows which time-based rules were applied
type TimeFactors struct {
	IsPeakHour         bool     `json:"is_peak_hour"`
	IsLateNight        bool     `json:"is_late_night"`
	IsWeekend          bool     `json:"is_weekend"`
	IsHoliday          bool     `json:"is_holiday"`
	IsPeakSeason       bool     `json:"is_peak_season"`
	IsSpecialEvent     bool     `json:"is_special_event"`
	SurgeActive        bool     `json:"surge_active"`
	SurgeLevel         float64  `json:"surge_level,omitempty"`
	CalculatedAt       string   `json:"calculated_at"`
	AppliedMultipliers []string `json:"applied_multipliers,omitempty"`
}

// CreateRuleRequest represents a request to create a new pricing rule with time-based and gemstone fields
type CreateRuleRequest struct {
	Name               string            `json:"name" binding:"required"`
	Strategy           string            `json:"strategy" binding:"required"`
	BasePrice          float64           `json:"base_price"`
	MarkupPercentage   float64           `json:"markup_percentage"`
	MinPrice           float64           `json:"min_price"`
	MaxPrice           float64           `json:"max_price"`
	DemandMultiplier   float64           `json:"demand_multiplier,omitempty"`
	RegionMultipliers  RegionMultipliers `json:"region_multipliers,omitempty"`
	DefaultCurrency    string            `json:"default_currency,omitempty"`
	TimeRules          TimeRules         `json:"time_rules,omitempty"`
	SurgeEnabled       bool              `json:"surge_enabled,omitempty"`
	BaseSurgeThreshold float64           `json:"base_surge_threshold,omitempty"`
	GemstoneConfig     GemstoneConfig    `json:"gemstone_config,omitempty"`
}

// Validate checks if the pricing rule is valid
func (r *CreateRuleRequest) Validate() error {
	validStrategies := map[string]bool{
		"cost_plus":    true,
		"demand_based": true,
		"competitive":  true,
		"geographic":   true,
		"time_based":   true,
		"gemstone":     true,
	}

	if !validStrategies[r.Strategy] {
		return errors.New("invalid strategy")
	}

	if r.MinPrice > r.MaxPrice && r.MaxPrice > 0 {
		return errors.New("min_price cannot be greater than max_price")
	}

	// Validate time-based strategy requirements
	if r.Strategy == "time_based" {
		if r.TimeRules.PeakMultiplier == 0 &&
			r.TimeRules.WeekendMultiplier == 0 &&
			r.TimeRules.HolidayMultiplier == 0 {
			return errors.New("time_based strategy requires at least one time rule")
		}
	}

	// Validate gemstone strategy requirements
	if r.Strategy == "gemstone" {
		if len(r.GemstoneConfig.BasePricePerCarat) == 0 {
			return errors.New("gemstone strategy requires base_price_per_carat")
		}
		if len(r.GemstoneConfig.CaratTiers) == 0 {
			return errors.New("gemstone strategy requires carat_tiers")
		}
		if len(r.GemstoneConfig.CutMultipliers) == 0 {
			return errors.New("gemstone strategy requires cut_multipliers")
		}
		if len(r.GemstoneConfig.ClarityMultipliers) == 0 {
			return errors.New("gemstone strategy requires clarity_multipliers")
		}
	}

	// Set defaults
	if r.DefaultCurrency == "" {
		r.DefaultCurrency = "USD"
	}
	if r.BaseSurgeThreshold == 0 {
		r.BaseSurgeThreshold = 1.0
	}

	return nil
}
