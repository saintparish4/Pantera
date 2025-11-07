-- 002_pricing_rules.down.sql
-- Rollback pricing_rules table

DROP TRIGGER IF EXISTS update_pricing_rules_updated_at ON pricing_rules;
DROP TABLE IF EXISTS pricing_rules;