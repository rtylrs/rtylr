-- ═══════════════════════════════════════════════════════════════════
-- POS V3 Schema: Paddle billing integration for POS orders
-- ═══════════════════════════════════════════════════════════════════

-- Store the Paddle transaction ID returned when a pay intent is created.
ALTER TABLE pos_orders
  ADD COLUMN IF NOT EXISTS paddle_transaction_id VARCHAR(64) DEFAULT NULL;

-- Track processed Paddle webhook events to enforce idempotent delivery.
CREATE TABLE IF NOT EXISTS pos_paddle_events_processed (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    transaction_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    owner_id BIGINT NOT NULL,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_pos_paddle_events_tx (transaction_id),
    INDEX idx_pos_paddle_events_owner (owner_id)
);
