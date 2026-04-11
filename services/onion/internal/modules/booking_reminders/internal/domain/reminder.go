package domain

import "time"

type Reminder struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	AppointmentID int64     `json:"appointmentId"`
	Type          string    `json:"type"`
	ScheduledAt   time.Time `json:"scheduledAt"`
	SentAt        time.Time `json:"sentAt"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ReminderInput struct {
	AppointmentID int64  `json:"appointmentId"`
	Type          string `json:"type"`
	ScheduledAt   string `json:"scheduledAt"`
}
