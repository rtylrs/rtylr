CREATE TABLE IF NOT EXISTS zontriq_publish_log (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES zontriq_videos(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES zontriq_channels(id) ON DELETE CASCADE,
    platform TEXT NOT NULL,
    platform_post_id TEXT,
    platform_post_url TEXT,
    result TEXT NOT NULL CHECK (result IN ('success', 'failed', 'removed')),
    error_message TEXT,
    published_at TIMESTAMPTZ DEFAULT NOW(),
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zontriq_publish_log_video_id ON zontriq_publish_log(video_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_publish_log_channel_id ON zontriq_publish_log(channel_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_publish_log_published_at ON zontriq_publish_log(published_at DESC);
