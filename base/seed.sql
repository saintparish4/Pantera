-- Seed data for testing Pantera API
-- Run with: psql -d pantera_db -f seed.sql

-- Example 1: Basic Cost-Plus Pricing
INSERT INTO pricing_rules (name, strategy, base_price, markup_percentage, min_price, max_price, active)
VALUES 
('Standard Product Pricing', 'cost_plus', 50.00, 30.0, 40.00, 100.00, true);

-- Example 2: Demand-Based Pricing (e.g., for concert tickets, hotels)
INSERT INTO pricing_rules (name, strategy, base_price, markup_percentage, min_price, max_price, demand_multiplier, active)
VALUES 
('Dynamic Event Pricing', 'demand_based', 100.00, 20.0, 80.00, 300.00, 0.5, true);

-- Example 3: Competitive Pricing (undercut competitors by 5%)
INSERT INTO pricing_rules (name, strategy, base_price, markup_percentage, min_price, max_price, active)
VALUES 
('Competitive Market Pricing', 'competitive', 75.00, -5.0, 60.00, 150.00, true);

-- Example 4: Premium Luxury Pricing
INSERT INTO pricing_rules (name, strategy, base_price, markup_percentage, min_price, max_price, active)
VALUES 
('Premium Product', 'cost_plus', 200.00, 50.0, 180.00, 400.00, true);

-- Example 5: Aggressive Demand-Based (e.g., Uber surge pricing)
INSERT INTO pricing_rules (name, strategy, base_price, markup_percentage, min_price, max_price, demand_multiplier, active)
VALUES 
('Surge Pricing', 'demand_based', 15.00, 10.0, 15.00, 75.00, 1.5, true);

-- Example 6: Inactive rule (for testing filtering)
INSERT INTO pricing_rules (name, strategy, base_price, markup_percentage, min_price, max_price, active)
VALUES 
('Deprecated Pricing', 'cost_plus', 30.00, 20.0, 25.00, 60.00, false);

-- Display inserted rules
SELECT id, name, strategy, base_price, markup_percentage, active 
FROM pricing_rules 
ORDER BY id;