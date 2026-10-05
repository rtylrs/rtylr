INSERT INTO authorization_permission_catalog
    (permission_key, app_id, action, is_sensitive, is_high_risk)
VALUES ('insights:settings:edit', 'insights', 'edit', FALSE, FALSE)
ON DUPLICATE KEY UPDATE
    app_id = VALUES(app_id),
    action = VALUES(action),
    is_sensitive = VALUES(is_sensitive),
    is_high_risk = VALUES(is_high_risk);

INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT id, 'insights:settings:edit'
FROM authorization_roles
WHERE business_id = 0 AND role_key = 'insights.admin' AND app_id = 'insights'
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);
