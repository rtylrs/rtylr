-- Paddle webhook integrity (#755): past_due is a real dunning state and must
# not collapse into "active", and delivered event ids need a durable dedupe
-- ledger so at-least-once redelivery cannot re-apply a stale transition.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'subscriptions' AND COLUMN_NAME = 'status' AND COLUMN_TYPE LIKE '%past_due%') = 0,
  'ALTER TABLE subscriptions MODIFY status ENUM(''active'',''trial'',''past_due'',''cancelled'',''expired'') DEFAULT ''trial''',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS paddle_webhook_events (
    event_id VARCHAR(255) NOT NULL PRIMARY KEY,
    paddle_subscription_id VARCHAR(255) NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_paddle_events_subscription_occurred (paddle_subscription_id, occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
