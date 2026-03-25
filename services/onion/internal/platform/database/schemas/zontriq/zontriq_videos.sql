CREATE TABLE IF NOT EXISTS zontriq_videos (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    series_id INTEGER REFERENCES zontriq_series(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    script TEXT,
    video_url TEXT,
    thumbnail_url TEXT,
    state TEXT NOT NULL DEFAULT 'draft' CHECK (state IN ('draft', 'generating', 'rendering', 'ready', 'scheduled', 'publishing', 'published', 'failed')),
    duration INTEGER,
    episode_number INTEGER,
    error_message TEXT,
    scheduled_publish_at TIMESTAMPTZ,
    config JSONB NOT NULL DEFAULT '{}',
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zontriq_videos_user_id ON zontriq_videos(zontriq_user_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_videos_series_id ON zontriq_videos(series_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_videos_state ON zontriq_videos(state);
