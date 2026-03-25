CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    contact_name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255),
    phone_number VARCHAR(50),
    email VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clients_email ON clients(email);
CREATE INDEX IF NOT EXISTS idx_clients_active ON clients(active);
CREATE INDEX IF NOT EXISTS idx_clients_company_name ON clients(company_name);
