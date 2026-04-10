package domain

import "time"

type SecuritySettings struct {
	TwoFactorEnabled bool            `json:"twoFactorEnabled"`
	ActiveSessions   []ActiveSession `json:"activeSessions"`
}

type ActiveSession struct {
	ID        string    `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
	IsCurrent bool      `json:"isCurrent"`
}

type TwoFactorInput struct {
	Enabled bool `json:"enabled"`
}
