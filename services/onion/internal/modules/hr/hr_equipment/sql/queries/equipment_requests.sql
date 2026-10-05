-- name: CreateEquipmentRequest :execresult
INSERT INTO hr_equipment_requests (owner_id, business_id, employee_id, item_name, quantity, reason, needed_by_date)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetEquipmentRequest :one
SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,
       current_approval_level, approved_by, rejection_reason, needed_by_date,
       created_at, updated_at
FROM hr_equipment_requests
WHERE business_id = ? AND id = ?;

-- name: ListEquipmentRequests :many
SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,
       current_approval_level, approved_by, rejection_reason, needed_by_date,
       created_at, updated_at
FROM hr_equipment_requests
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListEquipmentRequestsByEmployee :many
SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,
       current_approval_level, approved_by, rejection_reason, needed_by_date,
       created_at, updated_at
FROM hr_equipment_requests
WHERE business_id = ? AND employee_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListEquipmentRequestsByStatus :many
SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,
       current_approval_level, approved_by, rejection_reason, needed_by_date,
       created_at, updated_at
FROM hr_equipment_requests
WHERE business_id = ? AND status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListEquipmentRequestsByEmployeeAndStatus :many
SELECT id, owner_id, business_id, employee_id, item_name, quantity, reason, status,
       current_approval_level, approved_by, rejection_reason, needed_by_date,
       created_at, updated_at
FROM hr_equipment_requests
WHERE business_id = ? AND employee_id = ? AND status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountEquipmentRequests :one
SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ?;

-- name: CountEquipmentRequestsByEmployee :one
SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ? AND employee_id = ?;

-- name: CountEquipmentRequestsByStatus :one
SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ? AND status = ?;

-- name: CountEquipmentRequestsByEmployeeAndStatus :one
SELECT COUNT(*) AS total FROM hr_equipment_requests WHERE business_id = ? AND employee_id = ? AND status = ?;

-- name: UpdateEquipmentRequestStatus :execrows
UPDATE hr_equipment_requests
SET status = ?, current_approval_level = ?, approved_by = ?, rejection_reason = ?
WHERE business_id = ? AND id = ? AND status = 'pending' AND current_approval_level = ?;
