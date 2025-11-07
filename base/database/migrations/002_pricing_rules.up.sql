-- 002_pricing_rules.up.sql
-- Create pricing_rules table with JSONB config storage

CREATE TABLE pricing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    strategy_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_strategy_type CHECK (
        strategy_type IN ('cost_plus', 'geographic', 'time_based', 'rule_based')
    )
);

-- Indexes for pricing_rules
CREATE INDEX idx_pricing_rules_user ON pricing_rules(user_id, is_active, deleted_at);
CREATE INDEX idx_pricing_rules_strategy ON pricing_rules(strategy_type);
CREATE INDEX idx_pricing_rules_active ON pricing_rules(is_active, deleted_at);
CREATE INDEX idx_pricing_rules_config ON pricing_rules USING GIN (config);

-- Updated_at trigger
CREATE TRIGGER update_pricing_rules_updated_at BEFORE UPDATE ON pricing_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE pricing_rules IS 'Pricing rule configurations with flexible JSONB storage';
COMMENT ON COLUMN pricing_rules.strategy_type IS 'Pricing strategy: cost_plus, geographic, time_based, rule_based';
COMMENT ON COLUMN pricing_rules.config IS 'Strategy-specific configuration as JSONB';
COMMENT ON COLUMN pricing_rules.deleted_at IS 'Soft delete timestamp';

-- Sample data for development
INSERT INTO users (email) VALUES ('demo@harmonia.api') ON CONFLICT DO NOTHING;

-- Sample cost-plus rule
INSERT INTO pricing_rules (user_id, name, description, strategy_type, config)
SELECT 
    id,
    'Standard Cost-Plus',
    'Basic 25% markup on cost',
    'cost_plus',
    '{"markup_type": "percentage", "markup_value": 25, "tax_rate": 0.08}'::jsonb
FROM users WHERE email = 'demo@harmonia.api';

-- Sample geographic rule
INSERT INTO pricing_rules (user_id, name, description, strategy_type, config)
SELECT 
    id,
    'Regional Pricing - US',
    'US regional pricing with state multipliers',
    'geographic',
    '{
        "regional_multipliers": {
            "US": 1.0,
            "US-CA": 1.15,
            "US-NY": 1.20,
            "US-TX": 0.95
        },
        "default_currency": "USD"
    }'::jsonb
FROM users WHERE email = 'demo@harmonia.api';

-- Sample time-based rule
INSERT INTO pricing_rules (user_id, name, description, strategy_type, config)
SELECT 
    id,
    'Peak Hours Surge',
    'Surge pricing during peak hours',
    'time_based',
    '{
        "time_windows": [
            {
                "days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
                "start_time": "17:00",
                "end_time": "20:00",
                "multiplier": 1.5
            },
            {
                "days": ["saturday", "sunday"],
                "start_time": "12:00",
                "end_time": "22:00",
                "multiplier": 1.3
            }
        ]
    }'::jsonb
FROM users WHERE email = 'demo@harmonia.api';