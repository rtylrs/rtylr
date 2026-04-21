-- P3 — Add 'converted' status to finance_quotes so a quote marked as
-- materialised into an invoice can be distinguished from sent/accepted.
-- Idempotent: running MODIFY COLUMN on an enum that already contains the
-- value is a no-op.

ALTER TABLE finance_quotes
    MODIFY COLUMN status ENUM('draft','sent','accepted','rejected','expired','converted') DEFAULT 'draft';
