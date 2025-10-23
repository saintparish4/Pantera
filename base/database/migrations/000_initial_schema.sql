-- Migration: Initial schema setup
-- Creates base tables for the Pantera pricing engine

-- Create pricing_rules table
CREATE TABLE IF NOT EXISTS pricing_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    strategy VARCHAR(50) NOT NULL,
    base_price DECIMAL(10,2),
    markup_percentage DECIMAL(5,2),
    min_price DECIMAL(10,2),
    max_price DECIMAL(10,2),
    demand_multiplier DECIMAL(5,2) DEFAULT 1.0,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create price_calculations table (audit log)
CREATE TABLE IF NOT EXISTS price_calculations (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES pricing_rules(id),
    input_data JSONB,
    calculated_price DECIMAL(10,2),
    strategy_used VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_pricing_rules_active ON pricing_rules(active);
CREATE INDEX IF NOT EXISTS idx_price_calculations_rule_id ON price_calculations(rule_id);


