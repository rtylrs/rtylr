-- Durable analytics for Caret's exact-response cache.
-- The cache entries themselves may remain in Redis/local cache, but the console
-- needs hit/miss/write counters to survive API redeploys.
CREATE TABLE IF NOT EXISTS ai_cache_events (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(96) NOT NULL,
    occurred_at DATETIME(3) NOT NULL,
    outcome VARCHAR(32) NOT NULL,
    mode VARCHAR(32) NOT NULL,
    provider VARCHAR(100) NULL,
    reason VARCHAR(255) NULL,
    cache_key_preview VARCHAR(32) NULL,
    business_id BIGINT NULL,
    user_id BIGINT NULL,
    prompt_chars INT NULL,
    response_chars INT NULL,
    provider_tokens INT NULL,
    estimated_savings_tokens INT NULL,
    latency_ms BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_ai_cache_events_event_id (event_id),
    INDEX idx_ai_cache_events_occurred_at (occurred_at),
    INDEX idx_ai_cache_events_business_id (business_id),
    INDEX idx_ai_cache_events_outcome (outcome)
);
