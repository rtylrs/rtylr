











UPDATE flow_projects p
LEFT JOIN flow_workspaces w ON p.workspace_id = w.id AND w.is_personal = TRUE
SET p.visibility = 'private', p.is_personal = (w.id IS NOT NULL)
WHERE p.is_personal = FALSE;




UPDATE flow_labels l
JOIN flow_workspaces w ON l.workspace_id = w.id
SET l.business_id = w.business_id
WHERE l.business_id = 0;
