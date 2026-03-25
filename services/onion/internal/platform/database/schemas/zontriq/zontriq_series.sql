CREATE TABLE IF NOT EXISTS zontriq_series (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    niche TEXT NOT NULL,
    voice TEXT NOT NULL,
    style TEXT,
    frequency TEXT NOT NULL DEFAULT 'manual' CHECK (frequency IN ('manual', 'daily-1', 'daily-2', 'weekly-1', 'weekly-3')),
    state TEXT NOT NULL DEFAULT 'active' CHECK (state IN ('active', 'paused', 'archived')),
    video_count INTEGER NOT NULL DEFAULT 0,
    config JSONB NOT NULL DEFAULT '{}',
    publish_time TEXT,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zontriq_series_user_id ON zontriq_series(zontriq_user_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_series_state ON zontriq_series(state);
