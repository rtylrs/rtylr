package domain

import "time"

type Reminder struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	AppointmentID int64     `json:"appointment_id"`
	Type          string    `json:"type"`
	ScheduledAt   time.Time `json:"scheduled_at"`
	SentAt        time.Time `json:"sent_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type ReminderInput struct {
	AppointmentID int64  `json:"appointment_id"`
	Type          string `json:"type"`
	ScheduledAt   string `json:"scheduled_at"`
}
