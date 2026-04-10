-- name: ListJournalEntries :many
SELECT id, owner_id, date, description, created_at
FROM finance_journal_entries
WHERE owner_id = ?
ORDER BY date DESC;

-- name: CreateJournalEntry :execresult
INSERT INTO finance_journal_entries (owner_id, date, description)
VALUES (?, ?, ?);

-- name: GetJournalLines :many
SELECT id, entry_id, account_id, debit, credit
FROM finance_journal_lines
WHERE entry_id = ?
ORDER BY id ASC;

-- name: CreateJournalLine :exec
INSERT INTO finance_journal_lines (entry_id, account_id, debit, credit)
VALUES (?, ?, ?, ?);
