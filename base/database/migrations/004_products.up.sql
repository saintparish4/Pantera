-- 004_products.up.sql
-- Create products table for SKU management

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_cost DECIMAL(10,2),
    default_rule_id UUID REFERENCES pricing_rules(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_base_cost CHECK (base_cost >= 0),
    UNIQUE(user_id, sku)
);

-- Indexes for products
CREATE INDEX idx_products_user ON products(user_id, is_active);
CREATE INDEX idx_products_sku ON products(user_id, sku);
CREATE INDEX idx_products_rule ON products(default_rule_id);
CREATE INDEX idx_products_metadata ON products USING GIN (metadata);

-- Updated_at trigger
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE products IS 'Product/SKU catalog with base costs and default pricing rules';
COMMENT ON COLUMN products.sku IS 'Stock Keeping Unit - unique per user';
COMMENT ON COLUMN products.base_cost IS 'Base cost for cost-plus calculations';
COMMENT ON COLUMN products.default_rule_id IS 'Default pricing rule for this product';
COMMENT ON COLUMN products.metadata IS 'Additional product attributes (JSONB)';

-- Sample products for demo user
INSERT INTO products (user_id, sku, name, description, base_cost, metadata)
SELECT 
    id,
    'WIDGET-001',
    'Standard Widget',
    'Basic widget product',
    50.00,
    '{"category": "widgets", "weight_kg": 1.5, "dimensions": {"length": 10, "width": 5, "height": 3}}'::jsonb
FROM users WHERE email = 'demo@harmonia.api';

INSERT INTO products (user_id, sku, name, description, base_cost, metadata)
SELECT 
    id,
    'PREMIUM-001',
    'Premium Widget',
    'High-quality premium widget',
    120.00,
    '{"category": "widgets", "tier": "premium", "weight_kg": 2.0}'::jsonb
FROM users WHERE email = 'demo@harmonia.api';

INSERT INTO products (user_id, sku, name, description, base_cost, metadata)
SELECT 
    id,
    'SERVICE-CONSULT',
    'Consulting Service',
    'Hourly consulting service',
    150.00,
    '{"category": "services", "unit": "hour", "expertise_level": "senior"}'::jsonb
FROM users WHERE email = 'demo@harmonia.api';