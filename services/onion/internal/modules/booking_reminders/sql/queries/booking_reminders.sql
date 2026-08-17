-- name: ListReminders :many
SELECT id, owner_id, business_id, appointment_id, type, scheduled_at, sent_at, created_at
FROM booking_reminders WHERE business_id = ? ORDER BY scheduled_at DESC LIMIT 100;

-- name: CreateReminder :execresult
INSERT INTO booking_reminders (owner_id, business_id, appointment_id, type, scheduled_at) VALUES (?, ?, ?, ?, ?);

-- name: CountAppointmentForOwner :one
SELECT COUNT(*) FROM booking_appointments WHERE id = ? AND business_id = ?;
