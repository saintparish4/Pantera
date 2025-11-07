-- 003_calculation_logs.down.sql
-- Rollback calculation_logs table

DROP VIEW IF EXISTS calculation_analytics;
DROP TABLE IF EXISTS calculation_logs;