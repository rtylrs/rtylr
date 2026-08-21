-- name: ListSegments :many
SELECT id, owner_id, business_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at
FROM crm_segments WHERE business_id = ? ORDER BY created_at DESC LIMIT 100;

-- name: GetSegmentByID :one
SELECT id, owner_id, business_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at
FROM crm_segments WHERE id = ? AND business_id = ?;

-- name: CreateSegment :execresult
INSERT INTO crm_segments (owner_id, business_id, name, description, rules) VALUES (?, ?, ?, ?, ?);

-- name: UpdateSegment :execresult
UPDATE crm_segments SET name = ?, description = ?, rules = ? WHERE id = ? AND business_id = ?;

-- name: DeleteSegment :execresult
DELETE FROM crm_segments WHERE id = ? AND business_id = ?;

-- name: CountSegmentCustomers :one
-- cross-schema read from crm_customers — segment membership isn't computed
-- from rules yet (see infrastructure/repository.go), so this counts all
-- active customers for the business, matching the pre-existing legacy
-- behavior in the customers module.
SELECT COUNT(*) FROM crm_customers WHERE business_id = ? AND is_active = TRUE;

-- name: ListSegmentCustomers :many
-- cross-schema read from crm_customers — see CountSegmentCustomers note above.
SELECT id, owner_id, first_name, last_name, email, phone, date_of_birth, gender,
       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
       loyalty_tier, tags, notes, is_active, created_at, updated_at
FROM crm_customers WHERE business_id = ? AND is_active = TRUE
ORDER BY created_at DESC LIMIT ? OFFSET ?;
