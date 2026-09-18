-- Immutable tenant-scoped audit events for candidate lifecycle changes.

CREATE TABLE IF NOT EXISTS careers_submission_events (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    submission_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    actor_id BIGINT NOT NULL DEFAULT 0,
    event_type VARCHAR(64) NOT NULL,
    from_status VARCHAR(32) NULL,
    to_status VARCHAR(32) NULL,
    metadata_json JSON NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_submission_events_submission (submission_id, created_at, id),
    INDEX idx_submission_events_business (business_id, created_at)
);
