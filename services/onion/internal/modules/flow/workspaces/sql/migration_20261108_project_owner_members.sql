-- New projects list their creator as an admin member. Existing projects only
-- gave their owner access implicitly, so their member list read "0 members".
-- The owner already acts as admin, so this row grants nothing new; a user who
-- already has a member row keeps the role it has.
INSERT IGNORE INTO flow_project_members (project_id, user_id, role)
SELECT id, owner_id, 'admin'
FROM flow_projects
WHERE is_personal = FALSE AND owner_id <> 0;
