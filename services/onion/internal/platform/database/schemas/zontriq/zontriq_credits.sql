CREATE TABLE IF NOT EXISTS zontriq_credits (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    balance INTEGER NOT NULL DEFAULT 0,
    total_purchased INTEGER NOT NULL DEFAULT 0,
    total_spent INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (zontriq_user_id)
);
