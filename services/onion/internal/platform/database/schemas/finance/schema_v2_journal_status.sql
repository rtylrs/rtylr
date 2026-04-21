-- Phase 2 — journal posting workflow.
--
-- Adds a draft/posted status + posted_at timestamp so callers can
-- transition entries from draft to posted via
-- POST /finance/journal/{id}/post.
--
-- migrate.sh runs with `mysql --force`, so a duplicate-column error on
-- re-runs is logged as a warning and migration continues. First run
-- applies cleanly.

ALTER TABLE finance_journal_entries ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';
ALTER TABLE finance_journal_entries ADD COLUMN posted_at TIMESTAMP NULL;
