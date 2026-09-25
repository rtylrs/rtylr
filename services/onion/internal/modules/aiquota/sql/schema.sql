-- AI quota tables (local schema for sqlc — production DDL lives in migrations).
CREATE TABLE IF NOT EXISTS ai_quota_settings (
    owner_id BIGINT NOT NULL PRIMARY KEY,
    mode     VARCHAR(20) NOT NULL DEFAULT 'pool',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ai_usage_monthly (
    owner_id      BIGINT NOT NULL,
    user_id       BIGINT NOT NULL,
    billing_month DATE   NOT NULL,
    tokens_used   BIGINT NOT NULL DEFAULT 0,
    requests      INT    NOT NULL DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (owner_id, user_id, billing_month)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
