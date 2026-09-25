package domain

import "time"

type Settings struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"owner_id"`
	BusinessID     int64     `json:"business_id"`
	DefaultConsent bool      `json:"default_consent"`
	FromEmail      string    `json:"from_email"`
	FromName       string    `json:"from_name"`
	SmsSenderID    string    `json:"sms_sender_id"`
	Timezone       string    `json:"timezone"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SettingsInput struct {
	OwnerID        int64
	BusinessID     int64
	DefaultConsent bool
	FromEmail      string
	FromName       string
	SmsSenderID    string
	Timezone       string
}
