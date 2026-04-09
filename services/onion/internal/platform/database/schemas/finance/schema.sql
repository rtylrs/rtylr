CREATE TABLE IF NOT EXISTS finance_accounts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    type ENUM('asset','liability','equity','revenue','expense') NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    balance DECIMAL(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_invoices (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    invoice_number VARCHAR(100) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_email VARCHAR(255),
    client_address TEXT,
    client_tax_id VARCHAR(100),
    status ENUM('draft','sent','viewed','partial','paid','overdue','cancelled','void') DEFAULT 'draft',
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    paid_date DATE,
    currency VARCHAR(10) DEFAULT 'USD',
    subtotal DECIMAL(10,2) DEFAULT 0,
    discount_type ENUM('fixed','percentage') DEFAULT 'fixed',
    discount_value DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_rate DECIMAL(5,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) DEFAULT 0,
    amount_paid DECIMAL(10,2) DEFAULT 0,
    amount_due DECIMAL(10,2) DEFAULT 0,
    notes TEXT,
    terms TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_invoice_items (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    invoice_id BIGINT NOT NULL,
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(10,3) NOT NULL DEFAULT 1,
    unit VARCHAR(50) DEFAULT 'unit',
    unit_price DECIMAL(10,2) NOT NULL,
    tax_rate DECIMAL(5,2) DEFAULT 0,
    discount_pct DECIMAL(5,2) DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,
    sort_order INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS finance_payments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    payment_date DATE NOT NULL,
    method ENUM('cash','bank_transfer','card','cheque','online','other') NOT NULL,
    reference VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_expense_categories (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(20),
    icon VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_expenses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    category_id BIGINT,
    description VARCHAR(500) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    expense_date DATE NOT NULL,
    vendor VARCHAR(255),
    payment_method ENUM('cash','card','bank_transfer','other') DEFAULT 'cash',
    receipt_url TEXT,
    is_tax_deductible BOOLEAN DEFAULT FALSE,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_tax_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    rate DECIMAL(5,2) NOT NULL,
    type ENUM('vat','sales_tax','gst','other') DEFAULT 'vat',
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_recurring_invoices (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    template_invoice_id BIGINT NOT NULL,
    frequency ENUM('weekly','monthly','quarterly','yearly') NOT NULL,
    next_invoice_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
