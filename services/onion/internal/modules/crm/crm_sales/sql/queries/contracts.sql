-- name: ListContracts :many
SELECT * FROM crm_contracts
WHERE business_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ?;

-- name: ListSignedContracts :many
-- Backs crm.sales.v1's signed filter (#839): a caller asking for billable
-- contracts must not page through drafts.
SELECT * FROM crm_contracts
WHERE business_id = ? AND signed_at IS NOT NULL
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountSignedContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND signed_at IS NOT NULL;

-- name: ListUnsignedContracts :many
SELECT * FROM crm_contracts
WHERE business_id = ? AND signed_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountUnsignedContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND signed_at IS NULL;

-- name: GetContract :one
SELECT * FROM crm_contracts WHERE id = ? AND business_id = ?;

-- name: GetDealContract :one
SELECT * FROM crm_contracts WHERE business_id = ? AND deal_id = ? ORDER BY id DESC LIMIT 1;

-- name: CountDealContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND deal_id = ?;

-- name: CreateContract :execresult
INSERT INTO crm_contracts (owner_id, business_id, lead_id, deal_id, plan, mrr, seats,
                           value, billing, start_date, end_date, customer_id,
                           signed_at, signed_by_name, closed_by, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateContract :exec
UPDATE crm_contracts
SET plan = ?, mrr = ?, seats = ?, value = ?, billing = ?, start_date = ?, end_date = ?,
    customer_id = ?, signed_at = ?, signed_by_name = ?, closed_by = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: LinkDealContractsCustomer :exec
UPDATE crm_contracts SET customer_id = sqlc.arg('customer_id')
WHERE crm_contracts.business_id = sqlc.arg('business_id')
  AND crm_contracts.deal_id IN (
      SELECT d.id FROM crm_deals d WHERE d.business_id = sqlc.arg('business_id') AND d.lead_id = sqlc.arg('lead_id'))
  AND crm_contracts.customer_id IS NULL;
