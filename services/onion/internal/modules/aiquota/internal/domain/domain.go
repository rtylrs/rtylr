package domain

import (
	"context"
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
	PlanSlug        string    `json:"plan_slug"`
	Mode            Mode      `json:"mode"`
	AllowanceTokens int64     `json:"allowance_tokens"`
	UsedTokens      int64     `json:"used_tokens"`
	RemainingTokens int64     `json:"remaining_tokens"`
	OverageTokens   int64     `json:"overage_tokens"`
	AllowOverage    bool      `json:"allow_overage"`
	Blocked         bool      `json:"blocked"`
	PeriodEnd       time.Time `json:"period_end"`
}

// Repository reads plan entitlements (cross-schema), the tenant's allocation
// mode, and tracks token usage per bucket.
type Repository interface {
	// ResolveOwnerID maps a caller's business_id to the canonical tenant id
	// (businesses.owner_id). It returns an error rather than falling back to
	// fallbackUserID: the id scopes quota reads and token-usage writes, so an
	// unverified owner must never be substituted.
	ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) (int64, error)
	GetEntitlement(ctx context.Context, ownerID int64) (Entitlement, error)
	GetMode(ctx context.Context, ownerID int64) (Mode, error)
	SetMode(ctx context.Context, ownerID int64, mode Mode) error
	GetUsedTokens(ctx context.Context, ownerID, bucketUserID int64, month time.Time) (int64, error)
	AddUsage(ctx context.Context, ownerID, bucketUserID int64, month time.Time, tokens int64) error
}
