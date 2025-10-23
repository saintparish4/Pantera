-- Migration: Add time-based/surge pricing support

-- Add time-based pricing columns to pricing_rules
ALTER TABLE pricing_rules
ADD COLUMN time_rules JSONB DEFAULT '{}',
ADD COLUMN surge_enabled BOOLEAN DEFAULT false,
ADD COLUMN base_surge_threshold DECIMAL(5,2) DEFAULT 1.0;

COMMENT ON COLUMN pricing_rules.time_rules IS 'JSON object defining time-based pricing multipliers';
COMMENT ON COLUMN pricing_rules.surge_enabled IS 'Enable dynamic surge pricing based on demand';
COMMENT ON COLUMN pricing_rules.base_surge_threshold IS 'Base multiplier threshold for surge pricing';

-- Add timestamp and surge info to price_calculations
ALTER TABLE price_calculations
ADD COLUMN calculated_at_hour INTEGER,
ADD COLUMN calculated_at_day_of_week INTEGER,
ADD COLUMN surge_multiplier DECIMAL(5,2);

COMMENT ON COLUMN price_calculations.calculated_at_hour IS 'Hour of day when price was calculated (0-23)';
COMMENT ON COLUMN price_calculations.calculated_at_day_of_week IS 'Day of week (0=Sunday, 6=Saturday)';
COMMENT ON COLUMN price_calculations.surge_multiplier IS 'Surge multiplier applied to this calculation';

-- Create indexes for time-based analytics
CREATE INDEX idx_price_calculations_hour ON price_calculations(calculated_at_hour);
CREATE INDEX idx_price_calculations_day ON price_calculations(calculated_at_day_of_week);
CREATE INDEX idx_price_calculations_surge ON price_calculations(surge_multiplier);

-- Sample time-based pricing rule: Ride-sharing service
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    time_rules,
    surge_enabled,
    base_surge_threshold,
    active
) VALUES (
    'City Ride - Dynamic',
    'time_based',
    15.00,
    0,
    8.00,
    75.00,
    '{
        "peak_hours": [7, 8, 9, 17, 18, 19],
        "peak_multiplier": 1.5,
        "late_night_hours": [22, 23, 0, 1, 2, 3],
        "late_night_multiplier": 1.3,
        "weekend_days": [0, 6],
        "weekend_multiplier": 1.2,
        "holiday_dates": ["2025-12-25", "2025-01-01", "2025-07-04"],
        "holiday_multiplier": 1.8,
        "special_events": [
            {
                "name": "New Years Eve",
                "date": "2025-12-31",
                "start_hour": 20,
                "end_hour": 3,
                "multiplier": 2.5
            }
        ]
    }'::jsonb,
    true,
    1.0,
    true
);

-- Sample time-based pricing rule: Restaurant delivery
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    time_rules,
    surge_enabled,
    base_surge_threshold,
    active
) VALUES (
    'Food Delivery - Peak Hours',
    'time_based',
    5.99,
    0,
    3.99,
    15.99,
    '{
        "peak_hours": [11, 12, 13, 18, 19, 20],
        "peak_multiplier": 1.4,
        "weekend_days": [0, 6],
        "weekend_multiplier": 1.2,
        "late_night_hours": [22, 23, 0, 1],
        "late_night_multiplier": 1.6
    }'::jsonb,
    true,
    1.2,
    true
);

-- Sample time-based pricing rule: Hotel room (seasonal)
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    time_rules,
    surge_enabled,
    base_surge_threshold,
    active
) VALUES (
    'Deluxe Room - Seasonal',
    'time_based',
    199.00,
    0,
    99.00,
    499.00,
    '{
        "peak_months": [6, 7, 8, 11, 12],
        "peak_month_multiplier": 1.6,
        "weekend_days": [5, 6],
        "weekend_multiplier": 1.3,
        "holiday_dates": ["2025-12-24", "2025-12-25", "2025-12-31", "2026-01-01"],
        "holiday_multiplier": 2.0
    }'::jsonb,
    false,
    1.0,
    true
);

-- Sample time-based pricing rule: Event tickets
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    time_rules,
    surge_enabled,
    base_surge_threshold,
    active
) VALUES (
    'Concert VIP Ticket',
    'time_based',
    150.00,
    0,
    150.00,
    500.00,
    '{
        "early_bird_days_before": 60,
        "early_bird_multiplier": 0.8,
        "last_minute_days_before": 7,
        "last_minute_multiplier": 1.5,
        "day_of_event_multiplier": 2.0
    }'::jsonb,
    true,
    1.3,
    true
);

-- Create a view for time-based pricing analytics
CREATE OR REPLACE VIEW time_based_pricing_analytics AS
SELECT 
    pr.name as rule_name,
    pc.calculated_at_hour as hour,
    pc.calculated_at_day_of_week as day_of_week,
    AVG(pc.calculated_price) as avg_price,
    AVG(pc.surge_multiplier) as avg_surge_multiplier,
    COUNT(*) as calculation_count,
    MIN(pc.calculated_price) as min_price,
    MAX(pc.calculated_price) as max_price
FROM price_calculations pc
JOIN pricing_rules pr ON pc.rule_id = pr.id
WHERE pr.strategy = 'time_based'
GROUP BY pr.name, pc.calculated_at_hour, pc.calculated_at_day_of_week
ORDER BY pr.name, pc.calculated_at_hour;