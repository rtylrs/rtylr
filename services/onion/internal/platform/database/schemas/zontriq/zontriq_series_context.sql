CREATE TABLE IF NOT EXISTS zontriq_series_context (
    id SERIAL PRIMARY KEY,
    series_id INTEGER NOT NULL REFERENCES zontriq_series(id) ON DELETE CASCADE,
    context TEXT NOT NULL DEFAULT '',
    topics_covered JSONB NOT NULL DEFAULT '[]',
    last_episode_number INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (series_id)
);
