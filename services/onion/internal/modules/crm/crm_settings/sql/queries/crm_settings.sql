-- name: GetSettings :one
SELECT id, owner_id, business_id, default_consent, from_email, from_name, sms_sender_id, timezone, created_at, updated_at
FROM crm_settings WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO crm_settings
  (owner_id, business_id, default_consent, from_email, from_name, sms_sender_id, timezone)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  default_consent = VALUES(default_consent),
  from_email = VALUES(from_email),
  from_name = VALUES(from_name),
  sms_sender_id = VALUES(sms_sender_id),
  timezone = VALUES(timezone);
