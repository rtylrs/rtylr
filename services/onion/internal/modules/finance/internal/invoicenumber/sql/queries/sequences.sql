-- name: EnsureInvoiceSequence :exec
INSERT INTO finance_invoice_sequences (business_id, prefix, last_sequence)
VALUES (?, ?, 0)
ON DUPLICATE KEY UPDATE last_sequence = last_sequence;

-- name: LockInvoiceSequence :one
SELECT last_sequence FROM finance_invoice_sequences
WHERE business_id = ? AND prefix = ? FOR UPDATE;

-- name: MaxInvoiceSequence :one
SELECT CAST(COALESCE(MAX(CAST(SUBSTRING(invoice_number, LENGTH(sqlc.arg(prefix)) + 1) AS UNSIGNED)), 0) AS SIGNED) AS sequence
FROM finance_invoices
WHERE business_id = ? AND invoice_number LIKE CONCAT(sqlc.arg(prefix), '%')
  AND SUBSTRING_INDEX(invoice_number, '-', -1) REGEXP '^[0-9]+$';

-- name: AdvanceInvoiceSequence :exec
UPDATE finance_invoice_sequences SET last_sequence = ?
WHERE business_id = ? AND prefix = ?;
