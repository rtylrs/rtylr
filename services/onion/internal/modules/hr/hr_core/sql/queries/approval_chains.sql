-- name: GetApprovalChain :one
SELECT id, owner_id, business_id, request_type, name, created_at, updated_at
FROM hr_approval_chains
WHERE business_id = ? AND request_type = ?;

-- name: GetApprovalChainLevels :many
SELECT id, owner_id, business_id, approval_chain_id, level_order, approver_type,
       approver_role_id, approver_employee_id, created_at
FROM hr_approval_chain_levels
WHERE business_id = ? AND approval_chain_id = ?
ORDER BY level_order ASC;

-- name: UpsertApprovalChain :execresult
INSERT INTO hr_approval_chains (owner_id, business_id, request_type, name)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE name = VALUES(name), updated_at = NOW();

-- name: DeleteApprovalChainLevels :exec
DELETE FROM hr_approval_chain_levels
WHERE business_id = ? AND approval_chain_id = ?;

-- name: CreateApprovalChainLevel :exec
INSERT INTO hr_approval_chain_levels (owner_id, business_id, approval_chain_id, level_order, approver_type, approver_role_id, approver_employee_id)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: CreateInviteToken :execresult
INSERT INTO hr_employee_invites (owner_id, business_id, employee_id, token, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetInviteByToken :one
SELECT id, owner_id, business_id, employee_id, token, expires_at, consumed_at, created_at
FROM hr_employee_invites
WHERE token = ? AND consumed_at IS NULL;

-- name: ConsumeInviteToken :execresult
UPDATE hr_employee_invites
SET consumed_at = NOW(), updated_at = NOW()
WHERE token = ? AND consumed_at IS NULL;
