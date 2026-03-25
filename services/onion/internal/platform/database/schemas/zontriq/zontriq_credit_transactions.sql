CREATE TABLE IF NOT EXISTS zontriq_credit_transactions (
    id SERIAL PRIMARY KEY,
    zontriq_user_id INTEGER NOT NULL REFERENCES zontriq_users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('purchase', 'spend', 'refund')),
    description TEXT,
    reference_type TEXT,
    reference_id INTEGER,
    status INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
