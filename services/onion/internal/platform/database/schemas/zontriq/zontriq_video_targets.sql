CREATE TABLE IF NOT EXISTS zontriq_video_targets (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES zontriq_videos(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES zontriq_channels(id) ON DELETE CASCADE,
    state TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'publishing', 'published', 'failed')),
    published_at TIMESTAMPTZ,
    platform_post_id TEXT,
    error_message TEXT,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (video_id, channel_id)
);

CREATE INDEX IF NOT EXISTS idx_zontriq_video_targets_video_id ON zontriq_video_targets(video_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_video_targets_channel_id ON zontriq_video_targets(channel_id);
