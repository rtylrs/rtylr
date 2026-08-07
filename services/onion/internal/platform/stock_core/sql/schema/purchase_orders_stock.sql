-- erp_purchase_orders gains a receiving location (authoritative DDL: erp/migrations/002_unified_stock.sql).
-- Must load after erp/schema.sql defines erp_purchase_orders.
ALTER TABLE erp_purchase_orders ADD COLUMN destination_location_id BIGINT NULL;
