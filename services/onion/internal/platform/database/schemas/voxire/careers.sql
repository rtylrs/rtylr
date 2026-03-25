CREATE TABLE IF NOT EXISTS careers (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100),
    location VARCHAR(100),
    employment_type VARCHAR(50), -- e.g., Full-time, Part-time, Contract
    work_mode VARCHAR(20),       -- e.g., remote, hybrid, onsite
    seniority VARCHAR(50),       -- e.g., junior, mid, senior
    details TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);