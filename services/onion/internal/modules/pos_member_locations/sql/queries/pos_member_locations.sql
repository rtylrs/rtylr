-- name: ListMemberLocations :many
SELECT id, business_id, user_id, location_id, can_manage_pos_users,
       can_manage_terminals, can_view_reports, created_at
FROM pos_member_locations WHERE business_id = ? AND user_id = ?;

-- name: ListLocationMembers :many
SELECT id, business_id, user_id, location_id, can_manage_pos_users,
       can_manage_terminals, can_view_reports, created_at
FROM pos_member_locations WHERE business_id = ? AND location_id = ?;

-- name: AssignMemberToLocation :execresult
INSERT INTO pos_member_locations (business_id, user_id, location_id,
    can_manage_pos_users, can_manage_terminals, can_view_reports)
VALUES (?, ?, ?, ?, ?, ?);

-- name: RemoveMemberFromLocation :exec
DELETE FROM pos_member_locations WHERE business_id = ? AND id = ?;

-- name: GetMyLocations :many
SELECT ml.location_id, l.name, l.address, l.phone
FROM pos_member_locations ml
JOIN pos_locations l ON l.id = ml.location_id
WHERE ml.business_id = ? AND ml.user_id = ?;
