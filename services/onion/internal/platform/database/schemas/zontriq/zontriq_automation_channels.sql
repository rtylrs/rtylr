CREATE TABLE IF NOT EXISTS zontriq_automation_channels (
    id SERIAL PRIMARY KEY,
    automation_id INTEGER NOT NULL REFERENCES zontriq_automations(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES zontriq_channels(id) ON DELETE CASCADE,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (automation_id, channel_id)
);

CREATE INDEX IF NOT EXISTS idx_zontriq_automation_channels_automation_id ON zontriq_automation_channels(automation_id);
