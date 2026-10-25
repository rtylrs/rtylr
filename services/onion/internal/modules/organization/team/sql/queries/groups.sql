-- name: CreateGroup :execresult
INSERT INTO user_groups (business_id, name, description, color)
VALUES (?, ?, ?, ?);

-- name: GetGroup :one
SELECT id, business_id, name, COALESCE(description,'') AS description, color, created_at, updated_at
FROM user_groups WHERE id = ? AND business_id = ?;

-- name: ListGroups :many
SELECT g.id, g.business_id, g.name, COALESCE(g.description,'') AS description, g.color,
       g.created_at, g.updated_at, COUNT(m.user_id) AS member_count
FROM user_groups g
LEFT JOIN user_group_members m ON m.group_id = g.id AND m.business_id = g.business_id
WHERE g.business_id = ?
GROUP BY g.id
ORDER BY g.created_at DESC
LIMIT 500;

-- name: UpdateGroup :execresult
UPDATE user_groups SET name = ?, description = ?, color = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: DeleteGroup :execresult
DELETE FROM user_groups WHERE id = ? AND business_id = ?;

-- name: AddGroupMember :exec
INSERT INTO user_group_members (group_id, user_id, business_id, added_by)
SELECT ?, ?, ?, ?
FROM user_groups g
JOIN authorization_business_members m ON m.business_id = g.business_id AND m.user_id = ? AND m.status = 'active'
WHERE g.id = ? AND g.business_id = ?
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), added_by = VALUES(added_by);

-- name: RemoveGroupMember :exec
DELETE FROM user_group_members WHERE group_id = ? AND user_id = ? AND business_id = ?;

-- name: ListGroupMembers :many
SELECT m.user_id, u.email, COALESCE(u.full_name,'') AS full_name, m.added_at
FROM user_group_members m
JOIN users u ON u.id = m.user_id
WHERE m.group_id = ? AND m.business_id = ?
ORDER BY m.added_at DESC
LIMIT 500;

-- name: ListUserGroups :many
SELECT g.id, g.name, g.color
FROM user_group_members m
JOIN user_groups g ON g.id = m.group_id AND g.business_id = m.business_id
WHERE m.user_id = ? AND m.business_id = ?
ORDER BY g.name
LIMIT 500;

-- unbounded-ok: batch lookup keyed by the group_ids slice, bounded by the caller.
-- name: ListUserGroupsForUsers :many
SELECT m.user_id, g.id, g.name, g.color
FROM user_group_members m
JOIN user_groups g ON g.id = m.group_id AND g.business_id = m.business_id
WHERE m.user_id IN (sqlc.slice('user_ids')) AND m.business_id = ?
ORDER BY g.name;
