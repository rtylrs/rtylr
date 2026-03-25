CREATE TABLE IF NOT EXISTS zontriq_automations (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    series_id INTEGER NOT NULL REFERENCES zontriq_series(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    frequency TEXT NOT NULL CHECK (frequency IN ('daily-1', 'daily-2', 'weekly-1', 'weekly-3')),
    state TEXT NOT NULL DEFAULT 'active' CHECK (state IN ('active', 'paused')),
    next_run TIMESTAMPTZ,
    last_run TIMESTAMPTZ,
    videos_generated INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zontriq_automations_user_id ON zontriq_automations(zontriq_user_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_automations_next_run ON zontriq_automations(next_run) WHERE state = 'active';
