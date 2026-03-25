CREATE TABLE IF NOT EXISTS zontriq_preferences (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    default_voice TEXT,
    default_niche TEXT,
    notify_video_ready BOOLEAN NOT NULL DEFAULT TRUE,
    notify_publish_success BOOLEAN NOT NULL DEFAULT TRUE,
    notify_publish_failure BOOLEAN NOT NULL DEFAULT TRUE,
    notify_quota_warning BOOLEAN NOT NULL DEFAULT TRUE,
    notify_weekly_report BOOLEAN NOT NULL DEFAULT FALSE,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (zontriq_user_id)
);
