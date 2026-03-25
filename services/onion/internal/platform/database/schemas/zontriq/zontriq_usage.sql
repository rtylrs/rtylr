CREATE TABLE IF NOT EXISTS zontriq_usage (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    videos_generated INTEGER NOT NULL DEFAULT 0,
    videos_published INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (zontriq_user_id, period_start)
);

CREATE INDEX IF NOT EXISTS idx_zontriq_usage_user_period ON zontriq_usage(zontriq_user_id, period_end DESC);
