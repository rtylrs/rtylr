-- Phase 3 — chart-of-accounts hierarchy.
--
-- Adds a nullable parent_id to finance_accounts so the chart of
-- accounts can be rendered as a tree by GET /finance/chart-of-accounts.
-- Also adds an account code column so the tree can expose the standard
-- accounting code alongside the name.
--
-- migrate.sh runs with `mysql --force`, so duplicate-column errors on
-- re-runs are logged as warnings and migration continues. First run
-- applies cleanly.

ALTER TABLE finance_accounts ADD COLUMN code VARCHAR(32) NULL;
ALTER TABLE finance_accounts ADD COLUMN parent_id BIGINT NULL;
CREATE INDEX idx_finance_accounts_parent ON finance_accounts (parent_id);
