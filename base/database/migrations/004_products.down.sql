-- 004_products.down.sql
-- Rollback products table

DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TABLE IF EXISTS products;