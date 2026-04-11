-- name: ListReminders :many
SELECT id, owner_id, appointment_id, type, scheduled_at, sent_at, created_at
FROM booking_reminders WHERE owner_id = ? ORDER BY scheduled_at DESC;

-- name: CreateReminder :execresult
INSERT INTO booking_reminders (owner_id, appointment_id, type, scheduled_at) VALUES (?, ?, ?, ?);
