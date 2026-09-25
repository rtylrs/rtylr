-- name: CreateGroup :execresult
INSERT INTO user_groups (owner_id, name, description, color)
VALUES (?, ?, ?, ?);

-- name: GetGroup :one
SELECT id, owner_id, name, COALESCE(description,'') AS description, color, created_at, updated_at
FROM user_groups WHERE id = ? AND owner_id = ?;

-- Bounded at 500 like every other list in this file. ListGroupsRequest carries
-- no limit or offset and ListGroupsResponse carries no total, so real paging
-- needs a contract change first; an unbounded read is the worse of the two.
-- name: ListGroups :many
SELECT g.id, g.owner_id, g.name, COALESCE(g.description,'') AS description, g.color, g.created_at, g.updated_at,
       COUNT(m.user_id) AS member_count
FROM user_groups g
LEFT JOIN user_group_members m ON m.group_id = g.id
WHERE g.owner_id = ?
GROUP BY g.id
ORDER BY g.created_at DESC
LIMIT 500;

-- name: UpdateGroup :execresult
UPDATE user_groups SET name = ?, description = ?, color = ?, updated_at = NOW()
WHERE id = ? AND owner_id = ?;

-- name: DeleteGroup :execresult
DELETE FROM user_groups WHERE id = ? AND owner_id = ?;

-- name: UpsertGroupResourcePermission :exec
INSERT INTO user_group_resource_permissions (group_id, resource_key, permission_level)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE permission_level = VALUES(permission_level);

-- name: DeleteGroupResourcePermissions :exec
DELETE FROM user_group_resource_permissions WHERE group_id = ?;

-- name: ListGroupResourcePermissions :many
SELECT resource_key, permission_level
FROM user_group_resource_permissions
WHERE group_id = ?
LIMIT 500;

-- name: AddGroupMember :exec
-- business_id must be the caller's business, not the column default 0: the
-- authorization snapshot readers resolve group grants with
-- `business_id = <member's business>`, so a 0 row silently drops the group's
-- licenses and role assignments for that member. ON DUPLICATE KEY UPDATE
-- repairs rows written before this was set.
INSERT INTO user_group_members (group_id, user_id, owner_id, business_id, added_by)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), added_by = VALUES(added_by);

-- name: RemoveGroupMember :exec
DELETE FROM user_group_members WHERE group_id = ? AND user_id = ?;

-- name: ListGroupMembers :many
SELECT m.user_id, u.email, COALESCE(u.full_name,'') AS full_name, m.added_at
FROM user_group_members m
JOIN users u ON u.id = m.user_id
WHERE m.group_id = ?
ORDER BY m.added_at DESC
LIMIT 500;

-- name: ListUserGroups :many
SELECT g.id, g.name, g.color
FROM user_group_members m
JOIN user_groups g ON g.id = m.group_id
WHERE m.user_id = ? AND m.owner_id = ?
ORDER BY g.name
LIMIT 500;

-- unbounded-ok: one user's own group permissions, bounded by how many groups
-- that user belongs to. Not a list endpoint.
-- name: ListUserGroupResourcePermissions :many
SELECT ugm.group_id, ugrp.resource_key, ugrp.permission_level
FROM user_group_members ugm
JOIN user_group_resource_permissions ugrp ON ugm.group_id = ugrp.group_id
WHERE ugm.user_id = ? AND ugm.owner_id = ?;

-- unbounded-ok: batch lookup keyed by the user_ids slice, so the caller sets
-- the bound.
-- name: ListUserGroupsForUsers :many
SELECT m.user_id, g.id, g.name, g.color
FROM user_group_members m
JOIN user_groups g ON g.id = m.group_id
WHERE m.user_id IN (sqlc.slice('user_ids')) AND m.owner_id = ?
ORDER BY g.name;

-- unbounded-ok: batch lookup keyed by the group_ids slice, so the caller sets
-- the bound.
-- name: ListGroupResourcePermissionsForGroups :many
SELECT group_id, resource_key, permission_level
FROM user_group_resource_permissions
WHERE group_id IN (sqlc.slice('group_ids'));
