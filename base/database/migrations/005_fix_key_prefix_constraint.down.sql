-- 005_fix_key_prefix_constraint.down.sql
-- Revert to original constraint

ALTER TABLE api_keys 
DROP CONSTRAINT IF EXISTS chk_key_prefix;

ALTER TABLE api_keys 
ADD CONSTRAINT chk_key_prefix 
CHECK (key_prefix ~ '^hm_(live|test)_[a-zA-Z0-9]{8}$');

COMMENT ON COLUMN api_keys.key_prefix IS 'First 20 chars of key for display (e.g., hm_live_abc12345)';

