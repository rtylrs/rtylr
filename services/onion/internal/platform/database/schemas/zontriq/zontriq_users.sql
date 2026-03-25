CREATE TABLE IF NOT EXISTS zontriq_users (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id INTEGER NOT NULL REFERENCES zontriq_plans(id),
    status INTEGER NOT NULL DEFAULT 1,
    onboarded BOOLEAN NOT NULL DEFAULT FALSE,
    trial_ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS idx_zontriq_users_user_id ON zontriq_users(user_id);
CREATE INDEX IF NOT EXISTS idx_zontriq_users_plan_id ON zontriq_users(plan_id);
