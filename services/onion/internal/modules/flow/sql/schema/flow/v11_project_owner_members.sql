-- Projects created before the owner was added as a member show "0 members"
-- and leave the owner off their own Members list. Personal My Tasks buckets
-- are excluded: they are never shared and never list members.
INSERT IGNORE INTO `flow_project_members` (`project_id`, `user_id`, `role`, `added_by`)
SELECT p.`id`, p.`owner_id`, 'admin', p.`owner_id`
FROM `flow_projects` p
WHERE p.`owner_id` > 0
  AND p.`is_personal` = FALSE;
