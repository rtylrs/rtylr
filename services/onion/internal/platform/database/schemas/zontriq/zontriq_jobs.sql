CREATE TABLE IF NOT EXISTS zontriq_jobs (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES zontriq_videos(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('generate', 'render', 'publish')),
    state TEXT NOT NULL DEFAULT 'queued' CHECK (state IN ('queued', 'running', 'completed', 'failed')),
    progress INTEGER NOT NULL DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    current_step TEXT,
    error_message TEXT,
    gpu_id TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zontriq_jobs_video_id ON zontriq_jobs(video_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_jobs_state ON zontriq_jobs(state) WHERE state IN ('queued', 'running');
