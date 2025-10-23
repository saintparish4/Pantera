-- Migration: Add geographic pricing support

-- Add new columns to pricing_rules table
ALTER TABLE pricing_rules
ADD COLUMN region_multipliers JSONB DEFAULT '{}',
ADD COLUMN default_currency VARCHAR(3) DEFAULT 'USD';

-- Update pricing_rules table comment
COMMENT ON COLUMN pricing_rules.region_multipliers IS 'JSON object with region codes as keys and price multipliers as values';
COMMENT ON COLUMN pricing_rules.default_currency IS 'ISO 4217 currency code (USD, EUR, GBP, etc.)';

-- Add location column to price_calculations for audit trail
ALTER TABLE price_calculations
ADD COLUMN location VARCHAR(10),
ADD COLUMN currency VARCHAR(3);

COMMENT ON COLUMN price_calculations.location IS 'Location code used for calculation (e.g., US, US-CA, GB)';
COMMENT ON COLUMN price_calculations.currency IS 'Currency used for the calculation';

-- Create index for faster location-based queries
CREATE INDEX idx_price_calculations_location ON price_calculations(location);
CREATE INDEX idx_price_calculations_currency ON price_calculations(currency);

-- Sample geographic pricing rule
INSERT INTO pricing_rules (
    name, 
    strategy, 
    base_price, 
    markup_percentage,
    min_price, 
    max_price,
    region_multipliers,
    default_currency,
    active
) VALUES (
    'Global SaaS Subscription',
    'geographic',
    100.00,
    0,
    10.00,
    150.00,
    '{
        "US": 1.0,
        "US-CA": 1.15,
        "US-NY": 1.12,
        "CA": 0.95,
        "GB": 0.95,
        "EU": 1.05,
        "DE": 1.08,
        "FR": 1.05,
        "IN": 0.30,
        "BR": 0.40,
        "AU": 1.10,
        "JP": 1.05,
        "SG": 0.90
    }'::jsonb,
    'USD',
    true
);

-- Sample geographic e-commerce pricing
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    region_multipliers,
    default_currency,
    active
) VALUES (
    'Premium Product - Regional',
    'geographic',
    49.99,
    0,
    15.00,
    75.00,
    '{
        "US": 1.0,
        "US-CA": 1.2,
        "US-NY": 1.18,
        "US-TX": 0.95,
        "GB": 0.92,
        "EU": 1.1,
        "IN": 0.25,
        "CN": 0.35,
        "MX": 0.45
    }'::jsonb,
    'USD',
    true
);