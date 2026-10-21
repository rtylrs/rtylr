-- name: GetSettings :one
SELECT id, business_id, auto_assign, sla_hours, business_hours, notification_email, created_at, updated_at
FROM support_settings WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO support_settings (business_id, auto_assign, sla_hours, business_hours, notification_email)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE auto_assign = VALUES(auto_assign), sla_hours = VALUES(sla_hours),
business_hours = VALUES(business_hours), notification_email = VALUES(notification_email);
