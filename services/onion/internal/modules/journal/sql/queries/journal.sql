-- name: ListJournalEntries :many
SELECT id, owner_id, business_id, date, description, status, posted_at, created_at
FROM finance_journal_entries
WHERE business_id = ?
ORDER BY date DESC
LIMIT ? OFFSET ?;

-- name: GetJournalEntry :one
SELECT id, owner_id, business_id, date, description, status, posted_at, created_at
FROM finance_journal_entries
WHERE id = ? AND business_id = ?;

-- name: CreateJournalEntry :execresult
INSERT INTO finance_journal_entries (owner_id, business_id, date, description)
VALUES (?, ?, ?, ?);

-- name: PostJournalEntry :execresult
UPDATE finance_journal_entries
SET status = 'posted', posted_at = CURRENT_TIMESTAMP
WHERE id = ? AND business_id = ? AND status = 'draft';

-- name: GetJournalLines :many
SELECT id, entry_id, account_id, debit, credit
FROM finance_journal_lines
WHERE entry_id = ?
ORDER BY id ASC;

-- name: GetJournalLinesForEntries :many
SELECT id, entry_id, account_id, debit, credit
FROM finance_journal_lines
WHERE entry_id IN (sqlc.slice('entry_ids'))
ORDER BY id ASC;

-- name: CreateJournalLine :exec
INSERT INTO finance_journal_lines (entry_id, account_id, debit, credit)
VALUES (?, ?, ?, ?);

-- name: CountJournalEntriesByStatus :many
SELECT status, COUNT(*) AS cnt
FROM finance_journal_entries
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY status;

-- name: SumJournalLinesByAccount :one
SELECT
    COALESCE(SUM(CAST(l.debit AS REAL)), 0) AS total_debit,
    COALESCE(SUM(CAST(l.credit AS REAL)), 0) AS total_credit
FROM finance_journal_lines l
JOIN finance_journal_entries e ON e.id = l.entry_id
WHERE e.business_id = ?
  AND l.account_id = ?
  AND e.status = 'posted';
