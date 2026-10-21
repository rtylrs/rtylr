-- Member access changes were kept only in authorization_audit_events before
-- team changes reached the activity log. Copy them in once so the Audit page
-- shows that history; NOT EXISTS keeps a re-run from duplicating entries.
INSERT INTO activity_log
  (owner_id, business_id, app, entity_type, entity_id, action, description, user_id, old_values, new_values, created_at)
SELECT
  e.business_id,
  e.business_id,
  'hub',
  'member',
  e.target_user_id,
  'member.access_updated',
  CONCAT('Updated access for ',
    CASE
      WHEN NULLIF(TRIM(u.full_name), '') IS NOT NULL AND NULLIF(TRIM(u.email), '') IS NOT NULL
        THEN CONCAT(TRIM(u.full_name), ' (', TRIM(u.email), ')')
      ELSE COALESCE(NULLIF(TRIM(u.full_name), ''), NULLIF(TRIM(u.email), ''), 'a member')
    END),
  NULLIF(e.actor_user_id, 0),
  e.before_json,
  e.after_json,
  e.created_at
FROM authorization_audit_events e
LEFT JOIN users u ON u.id = e.target_user_id
WHERE e.action = 'member.authorization.updated'
  AND e.business_id > 0
  AND e.target_user_id > 0
  AND NOT EXISTS (
    SELECT 1 FROM activity_log a
    WHERE a.business_id = e.business_id
      AND a.action = 'member.access_updated'
      AND a.entity_id = e.target_user_id
      AND a.created_at = e.created_at
  );
