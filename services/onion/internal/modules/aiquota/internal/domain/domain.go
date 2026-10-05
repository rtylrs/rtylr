package domain

import (
	"time"
)

// Mode is the per-tenant allocation policy for the monthly AI allowance.
type Mode string

const (
	// ModePool shares one allowance across the whole team (default).
	ModePool Mode = "pool"
	// ModePerPerson splits the allowance evenly across seats.
	ModePerPerson Mode = "per_person"
)

// Normalize returns a valid Mode, defaulting unknown values to ModePool.
func (m Mode) Normalize() Mode {
	if m == ModePerPerson {
		return ModePerPerson
	}
	return ModePool
}

// Entitlement is the plan-derived monthly AI allowance for a tenant, read from
// the subscriptions + plans tables.
type Entitlement struct {
	PlanSlug          string
	AICreditsPerMonth int
	SeatCount         int
	Status            string
}

// QuotaStatus is the computed view of a tenant/bucket's current-month AI usage.
type QuotaStatus struct {
	PlanSlug        string
	Mode            Mode
	AllowanceTokens int64
	UsedTokens      int64
	RemainingTokens int64
	OverageTokens   int64
	AllowOverage    bool
	Blocked         bool
	PeriodEnd       time.Time
}
