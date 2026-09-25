-- Phase 4 — bank feed transactions.
--
-- Holds raw bank-side transactions imported from a connector (Plaid,
-- MX, etc.). The real connector is post-launch; GET /finance/bank-feeds
-- and POST /finance/bank-feeds/sync ship now with a stub sync so the
-- frontend can wire up the screen against the final contract.
--
-- matched_expense_id / matched_invoice_id are populated by the
-- reconciliation module (schema_v5_reconciliation) once a bank txn is
-- matched against a finance_expenses row or a finance_invoices row.
--
-- external_id + owner_id are unique together so re-syncing the same
-- upstream transaction is idempotent.
--
-- migrate.sh runs with `mysql --force`, so a duplicate-table error on
-- re-runs is logged as a warning and migration continues. First run
-- applies cleanly.

CREATE TABLE IF NOT EXISTS finance_bank_feeds (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    bank_account_id BIGINT NOT NULL,
    amount_minor_units BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    occurred_at TIMESTAMP NOT NULL,
    description TEXT,
    external_id VARCHAR(255) NOT NULL,
    matched_expense_id BIGINT NULL,
    matched_invoice_id BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_finance_bank_feeds_external (owner_id, external_id),
    KEY idx_finance_bank_feeds_owner (owner_id),
    KEY idx_finance_bank_feeds_bank_account (bank_account_id),
    KEY idx_finance_bank_feeds_occurred (occurred_at)
);
