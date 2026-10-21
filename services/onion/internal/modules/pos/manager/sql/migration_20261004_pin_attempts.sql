CREATE TABLE IF NOT EXISTS pos_manager_pin_attempts (
    business_id BIGINT NOT NULL,
    failures INT NOT NULL DEFAULT 0,
    last_failure_at BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (business_id)
);
