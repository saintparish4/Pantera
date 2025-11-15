-- 005_fix_key_prefix_constraint.up.sql
-- Fix the key_prefix constraint to match actual key format (first 20 chars)

ALTER TABLE api_keys 
DROP CONSTRAINT IF EXISTS chk_key_prefix;

ALTER TABLE api_keys 
ADD CONSTRAINT chk_key_prefix 
CHECK (key_prefix ~ '^hm_(live|test)_[a-f0-9]{12}$');

COMMENT ON COLUMN api_keys.key_prefix IS 'First 20 chars of key for display (e.g., hm_test_abc123456789)';

