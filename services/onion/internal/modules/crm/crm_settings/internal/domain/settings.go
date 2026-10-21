package domain

import "time"

type Settings struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	DefaultConsent bool
	FromEmail      string
	FromName       string
	SmsSenderID    string
	Timezone       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
