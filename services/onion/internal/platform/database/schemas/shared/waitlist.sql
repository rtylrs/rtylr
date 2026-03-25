CREATE TABLE IF NOT EXISTS waitlist (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    source TEXT DEFAULT 'landing',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
