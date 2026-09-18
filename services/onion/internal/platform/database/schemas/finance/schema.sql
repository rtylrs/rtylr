CREATE TABLE IF NOT EXISTS finance_accounts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(32) NULL,
    parent_id BIGINT NULL,
    type ENUM('asset','liability','equity','revenue','expense') NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    balance DECIMAL(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_finance_accounts_parent (parent_id),
    CONSTRAINT chk_finance_accounts_balance_non_negative CHECK (balance >= 0)
);

-- Tenant-selected accounts used when posting payroll. Payroll must not invent
-- chart-of-accounts entries; the accounts module owns this configuration.
CREATE TABLE IF NOT EXISTS finance_payroll_account_mappings (
    business_id BIGINT NOT NULL PRIMARY KEY,
    expense_account_id BIGINT NOT NULL,
    payable_account_id BIGINT NOT NULL,
    deductions_payable_account_id BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_finance_payroll_mapping_expense (expense_account_id),
    INDEX idx_finance_payroll_mapping_payable (payable_account_id),
    INDEX idx_finance_payroll_mapping_deductions (deductions_payable_account_id)
);

CREATE TABLE IF NOT EXISTS finance_invoices (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    invoice_number VARCHAR(100) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_email VARCHAR(255),
    client_address TEXT,
    client_tax_id VARCHAR(100),
    customer_id BIGINT NULL,
    status ENUM('draft','sent','viewed','partial','paid','overdue','cancelled','void') DEFAULT 'draft',
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    paid_date DATE,
    reporting_date DATE GENERATED ALWAYS AS (COALESCE(paid_date, issue_date)) STORED,
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    -- idx_finance_invoices_business_created / _business_status_created and the
    -- FULLTEXT search index are added by zz_finance_search_indexes.sql;
    -- declaring them here as well fails a fresh migration run with
    -- duplicate-key errors. The same applies to uq_finance_invoices_business_number
    -- and _business_status_reporting_date: zz_finance_invoice_number_unique.sql
    -- and zz_finance_reporting_date_live_repair.sql own those.
    INDEX idx_finance_invoices_business_status_reporting_date (business_id, status, reporting_date)
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
    business_id BIGINT NOT NULL DEFAULT 0,
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
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(20),
    icon VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_expenses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
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
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    -- idx_finance_expenses_business_date / _business_category_date and the
    -- FULLTEXT search index are added by zz_finance_search_indexes.sql;
    -- declaring them here as well fails a fresh migration run with
    -- duplicate-key errors.
);

CREATE TABLE IF NOT EXISTS finance_tax_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL UNIQUE,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(100) NOT NULL DEFAULT '',
    rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    type ENUM('vat','sales_tax','gst','other') DEFAULT 'vat',
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    tax_name VARCHAR(100),
    tax_rate DECIMAL(5,2) DEFAULT 0,
    is_inclusive BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_recurring_invoices (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    template_invoice_id BIGINT NOT NULL,
    frequency ENUM('weekly','monthly','quarterly','yearly') NOT NULL,
    next_invoice_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_quotes (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    status ENUM('draft','sent','accepted','rejected','expired','converted') DEFAULT 'draft',
    items JSON DEFAULT (JSON_ARRAY()),
    subtotal DECIMAL(10,2) DEFAULT 0,
    tax DECIMAL(10,2) DEFAULT 0,
    discount DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) DEFAULT 0,
    valid_until DATE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS finance_bills (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    supplier_id BIGINT,
    supplier_name VARCHAR(255) NOT NULL DEFAULT '',
    purchase_order_id BIGINT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    due_date DATE NOT NULL,
    status ENUM('pending','paid','overdue') DEFAULT 'pending',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_finance_bills_business_due (business_id, due_date, id),
    INDEX idx_finance_bills_business_status (business_id, status)
);

CREATE TABLE IF NOT EXISTS finance_bank_accounts (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    institution VARCHAR(255),
    account_type VARCHAR(50),
    balance DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'USD',
    is_connected BOOLEAN DEFAULT FALSE,
    last_synced_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_finance_bank_accounts_business (business_id)
);

CREATE TABLE IF NOT EXISTS finance_journal_entries (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    date DATE NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    source_type VARCHAR(32) NULL,
    source_id BIGINT NULL,
    posted_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_finance_journal_entries_business_date (business_id, date, id),
    INDEX idx_finance_journal_entries_business_status (business_id, status),
    UNIQUE KEY uq_finance_journal_entries_source (business_id, source_type, source_id)
);

CREATE TABLE IF NOT EXISTS finance_journal_lines (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    entry_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    debit DECIMAL(15,2) DEFAULT 0,
    credit DECIMAL(15,2) DEFAULT 0
);
