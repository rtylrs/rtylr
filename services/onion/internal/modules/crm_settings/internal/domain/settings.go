package domain

import "time"

type Settings struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"ownerId"`
	DefaultConsent bool      `json:"defaultConsent"`
	FromEmail      string    `json:"fromEmail"`
	FromName       string    `json:"fromName"`
	SmsSenderID    string    `json:"smsSenderId"`
	Timezone       string    `json:"timezone"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type SettingsInput struct {
	OwnerID        int64
	DefaultConsent bool
	FromEmail      string
	FromName       string
	SmsSenderID    string
	Timezone       string
}
