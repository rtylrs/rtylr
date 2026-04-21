-- name: ListAccounts :many
SELECT id, owner_id, name, type, currency, balance, is_active, created_at
FROM finance_accounts
WHERE owner_id = ?
ORDER BY type ASC, name ASC;

-- name: ListAccountsForTree :many
SELECT id, owner_id, name, code, parent_id, type, currency, balance, is_active, created_at
FROM finance_accounts
WHERE owner_id = ?
ORDER BY type ASC, code ASC, name ASC;

-- name: CreateAccount :execresult
INSERT INTO finance_accounts (owner_id, name, type, currency, balance, is_active)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateAccount :execresult
UPDATE finance_accounts
SET name = ?, type = ?, currency = ?, balance = ?, is_active = ?
WHERE id = ? AND owner_id = ?;

-- name: DeleteAccount :execresult
DELETE FROM finance_accounts
WHERE id = ? AND owner_id = ?;
