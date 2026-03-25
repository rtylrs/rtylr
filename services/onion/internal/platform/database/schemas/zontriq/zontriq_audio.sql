CREATE TABLE IF NOT EXISTS zontriq_audio (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    script TEXT NOT NULL,
    voice_id TEXT NOT NULL,
    voice_name TEXT,
    provider TEXT NOT NULL,
    audio_url TEXT,
    duration_seconds INTEGER,
    state TEXT NOT NULL DEFAULT 'queued' CHECK (state IN ('queued', 'generating', 'ready', 'failed')),
    retry_count INTEGER NOT NULL DEFAULT 0,
    external_job_id TEXT,
    error_message TEXT,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zontriq_audio_user_id ON zontriq_audio(zontriq_user_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_audio_state ON zontriq_audio(state);
