-- Migration: Add gemstone pricing support

-- Add gemstone pricing columns to pricing_rules
ALTER TABLE pricing_rules
ADD COLUMN gemstone_config JSONB DEFAULT '{}';

COMMENT ON COLUMN pricing_rules.gemstone_config IS 'Configuration for gemstone pricing including base rates and multipliers';

-- Add gemstone attributes to price_calculations
ALTER TABLE price_calculations
ADD COLUMN gemstone_type VARCHAR(50),
ADD COLUMN carat_weight DECIMAL(10,2),
ADD COLUMN cut_grade VARCHAR(20),
ADD COLUMN clarity_grade VARCHAR(10),
ADD COLUMN color_grade VARCHAR(10),
ADD COLUMN certification VARCHAR(20);

COMMENT ON COLUMN price_calculations.gemstone_type IS 'Type of gemstone (diamond, ruby, emerald, etc.)';
COMMENT ON COLUMN price_calculations.carat_weight IS 'Weight of gemstone in carats';
COMMENT ON COLUMN price_calculations.cut_grade IS 'Cut quality grade';
COMMENT ON COLUMN price_calculations.clarity_grade IS 'Clarity grade';
COMMENT ON COLUMN price_calculations.color_grade IS 'Color grade';
COMMENT ON COLUMN price_calculations.certification IS 'Certification lab (GIA, IGI, etc.)';

-- Create indexes for gemstone queries
CREATE INDEX idx_price_calculations_gemstone_type ON price_calculations(gemstone_type);
CREATE INDEX idx_price_calculations_carat ON price_calculations(carat_weight);
CREATE INDEX idx_price_calculations_cut ON price_calculations(cut_grade);

-- Sample gemstone pricing rule - Luxury Jewelry Store
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    gemstone_config,
    active
) VALUES (
    'Premium Gemstones - Luxury Collection',
    'gemstone',
    0, -- Base price determined by gemstone type
    0,
    100.00,
    500000.00,
    '{
        "base_price_per_carat": {
            "diamond": 5000,
            "ruby": 3000,
            "emerald": 2500,
            "sapphire": 2000,
            "tanzanite": 1500,
            "amethyst": 400
        },
        "carat_tiers": {
            "0.5-1.0": 1.0,
            "1.0-2.0": 1.3,
            "2.0-3.0": 1.6,
            "3.0+": 2.0
        },
        "cut_multipliers": {
            "poor": 0.7,
            "good": 0.9,
            "very_good": 1.0,
            "excellent": 1.15,
            "ideal": 1.3
        },
        "clarity_multipliers": {
            "I1": 0.6,
            "SI2": 0.75,
            "SI1": 0.85,
            "VS2": 1.0,
            "VS1": 1.2,
            "VVS2": 1.4,
            "VVS1": 1.6,
            "IF": 1.8,
            "FL": 2.0
        },
        "color_multipliers": {
            "K-M": 0.7,
            "J": 0.85,
            "I": 0.9,
            "H": 0.95,
            "G": 1.0,
            "F": 1.1,
            "E": 1.2,
            "D": 1.3
        },
        "certification_multipliers": {
            "none": 0.85,
            "IGI": 0.95,
            "AGS": 1.0,
            "GIA": 1.1
        },
        "treatment_multipliers": {
            "treated": 0.7,
            "heated": 0.85,
            "natural": 1.0
        }
    }'::jsonb,
    true
);

-- Sample rule for online diamond marketplace (competitive pricing)
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    gemstone_config,
    active
) VALUES (
    'Online Diamond Marketplace',
    'gemstone',
    0,
    0,
    50.00,
    100000.00,
    '{
        "base_price_per_carat": {
            "diamond": 4200,
            "ruby": 2500,
            "emerald": 2000,
            "sapphire": 1600
        },
        "carat_tiers": {
            "0.5-1.0": 1.0,
            "1.0-2.0": 1.25,
            "2.0-3.0": 1.5,
            "3.0+": 1.8
        },
        "cut_multipliers": {
            "good": 0.85,
            "very_good": 0.95,
            "excellent": 1.0,
            "ideal": 1.2
        },
        "clarity_multipliers": {
            "SI2": 0.7,
            "SI1": 0.8,
            "VS2": 0.9,
            "VS1": 1.0,
            "VVS2": 1.25,
            "VVS1": 1.5,
            "IF": 1.7,
            "FL": 2.0
        },
        "color_multipliers": {
            "J": 0.8,
            "I": 0.85,
            "H": 0.9,
            "G": 0.95,
            "F": 1.0,
            "E": 1.1,
            "D": 1.2
        },
        "certification_multipliers": {
            "IGI": 0.95,
            "GIA": 1.0
        }
    }'::jsonb,
    true
);

-- Sample rule for colored gemstone specialist
INSERT INTO pricing_rules (
    name,
    strategy,
    base_price,
    markup_percentage,
    min_price,
    max_price,
    gemstone_config,
    active
) VALUES (
    'Colored Gemstone Gallery',
    'gemstone',
    0,
    0,
    200.00,
    250000.00,
    '{
        "base_price_per_carat": {
            "ruby": 3500,
            "emerald": 2800,
            "sapphire": 2200,
            "tanzanite": 1800,
            "amethyst": 500,
            "aquamarine": 800,
            "topaz": 300,
            "garnet": 250
        },
        "carat_tiers": {
            "0.5-1.0": 1.0,
            "1.0-2.0": 1.35,
            "2.0-5.0": 1.8,
            "5.0+": 2.2
        },
        "cut_multipliers": {
            "poor": 0.6,
            "good": 0.85,
            "very_good": 1.0,
            "excellent": 1.2
        },
        "clarity_multipliers": {
            "included": 0.65,
            "slightly_included": 0.8,
            "eye_clean": 1.0,
            "loupe_clean": 1.3,
            "flawless": 1.7
        },
        "color_multipliers": {
            "light": 0.7,
            "medium": 0.85,
            "vivid": 1.0,
            "deep": 1.15,
            "pigeon_blood": 1.5,
            "kashmir_blue": 1.6
        },
        "treatment_multipliers": {
            "treated": 0.6,
            "heated": 0.75,
            "oiled": 0.85,
            "natural": 1.0
        },
        "origin_multipliers": {
            "unknown": 0.9,
            "standard": 1.0,
            "burma_ruby": 1.4,
            "kashmir_sapphire": 1.6,
            "colombian_emerald": 1.3,
            "tanzanian_tanzanite": 1.1
        }
    }'::jsonb,
    true
);

-- Create view for gemstone pricing analytics
CREATE OR REPLACE VIEW gemstone_pricing_analytics AS
SELECT 
    pr.name as rule_name,
    pc.gemstone_type,
    AVG(pc.carat_weight) as avg_carat,
    AVG(pc.calculated_price) as avg_price,
    AVG(pc.calculated_price / pc.carat_weight) as avg_price_per_carat,
    COUNT(*) as calculation_count,
    MIN(pc.calculated_price) as min_price,
    MAX(pc.calculated_price) as max_price,
    pc.cut_grade,
    pc.clarity_grade,
    pc.color_grade
FROM price_calculations pc
JOIN pricing_rules pr ON pc.rule_id = pr.id
WHERE pr.strategy = 'gemstone'
  AND pc.carat_weight IS NOT NULL
GROUP BY pr.name, pc.gemstone_type, pc.cut_grade, pc.clarity_grade, pc.color_grade
ORDER BY avg_price DESC;

-- Create function to get popular gemstone configurations
CREATE OR REPLACE VIEW popular_gemstone_configs AS
SELECT 
    gemstone_type,
    carat_weight,
    cut_grade,
    clarity_grade,
    color_grade,
    COUNT(*) as search_count,
    AVG(calculated_price) as avg_price
FROM price_calculations
WHERE strategy_used = 'gemstone'
  AND gemstone_type IS NOT NULL
  AND created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY gemstone_type, carat_weight, cut_grade, clarity_grade, color_grade
HAVING COUNT(*) >= 3
ORDER BY search_count DESC
LIMIT 20;