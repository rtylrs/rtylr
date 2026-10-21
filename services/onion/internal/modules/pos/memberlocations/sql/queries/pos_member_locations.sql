-- name: ListMemberLocations :many
SELECT id, business_id, user_id, location_id, can_manage_pos_users,
       can_manage_terminals, can_view_reports, created_at
FROM pos_member_locations WHERE business_id = ? AND user_id = ?;

-- name: AssignMemberToLocation :execresult
-- Idempotent on (user_id, location_id) so re-assigning a member (e.g. role
-- change in dash) updates their management flags instead of erroring.
INSERT INTO pos_member_locations (business_id, user_id, location_id,
    can_manage_pos_users, can_manage_terminals, can_view_reports)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    id = LAST_INSERT_ID(id),
    can_manage_pos_users = VALUES(can_manage_pos_users),
    can_manage_terminals = VALUES(can_manage_terminals),
    can_view_reports = VALUES(can_view_reports);

-- name: RemoveMemberFromLocation :exec
DELETE FROM pos_member_locations WHERE business_id = ? AND id = ?;

-- cross-schema read from pos_locations — enrich member assignments with POS location details.
-- name: GetMyLocations :many
SELECT ml.location_id, l.name, l.address, l.phone
FROM pos_member_locations ml
JOIN pos_locations l ON l.id = ml.location_id AND l.business_id = ml.business_id
WHERE ml.business_id = ? AND ml.user_id = ?
ORDER BY l.name ASC;

-- cross-schema read from pos_locations — list active locations belonging to the selected business.
-- name: ListBusinessLocations :many
SELECT l.id AS location_id, l.name, l.address, l.phone
FROM pos_locations l
WHERE l.business_id = ? AND l.is_active = TRUE
ORDER BY l.name ASC;

-- cross-schema read from pos_users — enrich the location roster with operator PIN and role state.
-- name: GetLocationRoster :many
-- The operator sign-in roster for a location: every assigned team member, their
-- display name (from users), whether they already hold a PIN here, and their
-- POS role if so. cross-schema read from users + pos_users.
SELECT ml.user_id,
       ml.can_manage_pos_users,
       (pu.id IS NOT NULL) AS has_pin,
       COALESCE(pu.role, '') AS pos_role
FROM pos_member_locations ml
LEFT JOIN pos_users pu
       ON pu.user_id = ml.user_id
      AND pu.location_id = ml.location_id
      AND pu.business_id = ml.business_id
      AND pu.is_active = TRUE
WHERE ml.business_id = ? AND ml.location_id = ?
ORDER BY ml.user_id ASC;

-- cross-schema read from pos_locations — validate that the requested location belongs to the business.
-- name: LocationBelongsToBusiness :one
SELECT EXISTS (
    SELECT 1
    FROM pos_locations l
    WHERE l.business_id = ? AND l.id = ?
);
