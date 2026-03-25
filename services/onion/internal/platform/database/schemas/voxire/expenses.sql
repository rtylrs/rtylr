CREATE TABLE IF NOT EXISTS expenses (
    id SERIAL PRIMARY KEY,
    product VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    term VARCHAR(100),
    price DECIMAL(10,2) NOT NULL,
    payment_date DATE,
    paid BOOLEAN NOT NULL DEFAULT false,
    extra_info TEXT,
    subscription BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);