-- ═══════════════════════════════════════════════════════════════════
-- POS V4 Schema: Store customer phone on online orders
-- ═══════════════════════════════════════════════════════════════════

ALTER TABLE pos_orders
  ADD COLUMN customer_phone VARCHAR(50) DEFAULT NULL AFTER customer_name;
