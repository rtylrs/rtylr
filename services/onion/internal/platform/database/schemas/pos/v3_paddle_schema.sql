-- ═══════════════════════════════════════════════════════════════════
-- POS V3 Schema: Paddle billing integration for POS orders
-- ═══════════════════════════════════════════════════════════════════

-- paddle_transaction_id is defined in the consolidated base schema.
-- Keeping it there avoids duplicate-column errors in schema generation.

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
