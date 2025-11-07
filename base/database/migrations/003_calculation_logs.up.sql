-- 003_calculation_logs.up.sql
-- Create calculation_logs table for audit trail

CREATE TABLE calculation_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    rule_id UUID REFERENCES pricing_rules(id) ON DELETE SET NULL,
    strategy_type VARCHAR(50) NOT NULL,
    input_data JSONB NOT NULL,
    output_data JSONB NOT NULL,
    execution_time_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_execution_time CHECK (execution_time_ms >= 0)
);

-- Indexes for calculation_logs (optimized for common queries)
CREATE INDEX idx_calc_logs_user_created ON calculation_logs(user_id, created_at DESC);
CREATE INDEX idx_calc_logs_api_key_created ON calculation_logs(api_key_id, created_at DESC);
CREATE INDEX idx_calc_logs_rule ON calculation_logs(rule_id);
CREATE INDEX idx_calc_logs_strategy ON calculation_logs(strategy_type);
CREATE INDEX idx_calc_logs_created ON calculation_logs(created_at DESC);

-- GIN indexes for JSONB queries
CREATE INDEX idx_calc_logs_input ON calculation_logs USING GIN (input_data);
CREATE INDEX idx_calc_logs_output ON calculation_logs USING GIN (output_data);

-- Partitioning by month (optional, for high-volume scenarios)
-- This can be added later if needed

-- Comments
COMMENT ON TABLE calculation_logs IS 'Audit trail of all price calculations';
COMMENT ON COLUMN calculation_logs.input_data IS 'Input parameters sent to pricing engine';
COMMENT ON COLUMN calculation_logs.output_data IS 'Calculated price with full breakdown';
COMMENT ON COLUMN calculation_logs.execution_time_ms IS 'Calculation execution time in milliseconds';

-- View for calculation analytics
CREATE VIEW calculation_analytics AS
SELECT 
    user_id,
    strategy_type,
    COUNT(*) as total_calculations,
    AVG(execution_time_ms) as avg_execution_time_ms,
    MIN((output_data->>'final_price')::numeric) as min_price,
    MAX((output_data->>'final_price')::numeric) as max_price,
    AVG((output_data->>'final_price')::numeric) as avg_price,
    DATE_TRUNC('day', created_at) as calculation_date
FROM calculation_logs
WHERE created_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
GROUP BY user_id, strategy_type, DATE_TRUNC('day', created_at);

COMMENT ON VIEW calculation_analytics IS 'Last 30 days of calculation statistics by user and strategy';