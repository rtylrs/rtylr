-- Repairs POS order/session runtime columns expected by generated queries.
-- Duplicate-column/index errors are safe under the migration runner's --force mode.

ALTER TABLE pos_sessions ADD COLUMN business_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE pos_sessions ADD COLUMN terminal_id BIGINT DEFAULT NULL;

ALTER TABLE pos_orders MODIFY COLUMN status
  ENUM('pending','confirmed','preparing','ready','served','paid','cancelled','refunded','voided','held')
  DEFAULT 'pending';
ALTER TABLE pos_orders ADD COLUMN terminal_id BIGINT DEFAULT NULL;
ALTER TABLE pos_orders ADD COLUMN branch_id BIGINT DEFAULT NULL;
ALTER TABLE pos_orders ADD COLUMN tip_amount DECIMAL(10,2) DEFAULT 0;
ALTER TABLE pos_orders ADD COLUMN void_reason TEXT DEFAULT NULL;
ALTER TABLE pos_orders ADD COLUMN voided_by BIGINT DEFAULT NULL;
ALTER TABLE pos_orders ADD COLUMN voided_at TIMESTAMP NULL;
ALTER TABLE pos_orders ADD COLUMN exchange_order_id BIGINT DEFAULT NULL;
ALTER TABLE pos_orders ADD COLUMN paddle_transaction_id VARCHAR(64) DEFAULT NULL;
ALTER TABLE pos_orders ADD INDEX idx_pos_orders_paddle_transaction (paddle_transaction_id);
