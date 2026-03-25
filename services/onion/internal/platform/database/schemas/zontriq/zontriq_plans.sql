CREATE TABLE IF NOT EXISTS zontriq_plans (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    price_cents INTEGER NOT NULL DEFAULT 0,
    max_videos_per_month INTEGER NOT NULL DEFAULT 0,
    max_channels INTEGER NOT NULL DEFAULT 0,
    max_series INTEGER NOT NULL DEFAULT 0,
    max_automations INTEGER NOT NULL DEFAULT 0,
    features JSONB NOT NULL DEFAULT '{}',
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO zontriq_plans (name, display_name, price_cents, max_videos_per_month, max_channels, max_series, max_automations) VALUES
    ('starter', 'Starter', 900, 10, 2, 3, 1),
    ('creator', 'Creator', 1900, 30, 5, 10, 5),
    ('pro', 'Pro', 3900, 100, 15, 50, 20),
    ('agency', 'Agency', 9900, 500, 50, 200, 100)
ON CONFLICT (name) DO NOTHING;
