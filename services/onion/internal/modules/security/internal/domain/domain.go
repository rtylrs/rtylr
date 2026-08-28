package domain

import "time"

type SecuritySettings struct {
	TwoFactorEnabled bool            `json:"two_factor_enabled"`
	MFARequired      bool            `json:"mfa_required"`
	MFAEnrolled      bool            `json:"mfa_enrolled"`
	ActiveSessions   []ActiveSession `json:"active_sessions"`
}

type ActiveSession struct {
	ID        string    `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	IsCurrent bool      `json:"is_current"`
}

type TwoFactorInput struct {
	Enabled bool `json:"enabled"`
}

type LoginEntry struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Method    string    `json:"method"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}
