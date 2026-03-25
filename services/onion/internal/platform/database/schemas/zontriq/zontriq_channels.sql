CREATE TABLE IF NOT EXISTS zontriq_channels (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    platform TEXT NOT NULL CHECK (platform IN ('youtube', 'tiktok', 'instagram')),
    platform_channel_id TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    channel_handle TEXT,
    channel_avatar_url TEXT,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ,
    connection_state TEXT NOT NULL DEFAULT 'connected' CHECK (connection_state IN ('connected', 'token_expired', 'revoked', 'disconnected')),
    connected_at TIMESTAMPTZ DEFAULT NOW(),
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (zontriq_user_id, platform, platform_channel_id)
);

CREATE INDEX IF NOT EXISTS idx_zontriq_channels_user_id ON zontriq_channels(zontriq_user_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_channels_connection_state ON zontriq_channels(connection_state);
