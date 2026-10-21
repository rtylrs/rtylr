package domain

import (
	"errors"
	"time"
)

type RecoveryCodeRow struct {
	ID   int64
	Hash string
}

var ErrMFAEnforcedByOrganization = errors.New("MFA is required by the organization")

type SecuritySettings struct {
	TwoFactorEnabled bool
	MFARequired      bool
	MFAEnrolled      bool
	ActiveSessions   []ActiveSession
}

type ActiveSession struct {
	ID        string
	IP        string
	UserAgent string
	CreatedAt time.Time
	IsCurrent bool
}

// MFAFactorState is what a step-up decision is made from: whether the user
// currently holds an active second factor, and when they last proved it.
// LastVerifiedAt is nil when nothing has ever been verified, which is not the
// same thing as a failed read.
type MFAFactorState struct {
	FactorEnrolled bool
	LastVerifiedAt *time.Time
}
type LoginEntry struct {
	ID        int64
	Email     string
	Method    string
	IP        string
	UserAgent string
	Success   bool
	CreatedAt time.Time
}
