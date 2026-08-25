package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

const FreeIncludedSeatLimit = 10

type Plan struct {
	ID                      int64     `json:"id"`
	Name                    string    `json:"name"`
	Slug                    string    `json:"slug"`
	Apps                    []string  `json:"apps"`
	PricePerSeatCents       int       `json:"price_per_seat_cents"`
	AnnualPricePerSeatCents int       `json:"annual_price_per_seat_cents"`
	TrialDays               int       `json:"trial_days"`
	PaddlePriceID           string    `json:"paddle_price_id,omitempty"`
	Tagline                 string    `json:"tagline"`
	MaxSeats                *int      `json:"max_seats"`
	MaxLocations            *int      `json:"max_locations"`
	AICreditsPerMonth       int       `json:"ai_credits_per_month"`
	Features                []string  `json:"features"`
	TierOrder               int       `json:"tier_order"`
	IsSolo                  bool      `json:"is_solo"`
	IsFeatured              bool      `json:"is_featured"`
	IsActive                bool      `json:"is_active"`
	CreatedAt               time.Time `json:"created_at"`
}

type Subscription struct {
	ID                   int64      `json:"id"`
	OwnerID              int64      `json:"owner_id"`
	PlanID               int64      `json:"plan_id"`
	PlanName             string     `json:"plan_name"`
	PlanSlug             string     `json:"plan_slug"`
	Apps                 []string   `json:"apps"`
	PricePerSeatCents    int        `json:"price_per_seat_cents"`
	Status               string     `json:"status"`
	IsGranted            bool       `json:"is_granted"`
	SeatCount            int        `json:"seat_count"`
	IncludedSeatLimit    int        `json:"included_seat_limit"`
	IncludedSeatCount    int        `json:"included_seat_count"`
	BillableSeatCount    int        `json:"billable_seat_count"`
	GrantExpiresAt       *time.Time `json:"grant_expires_at,omitempty"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
	NextBillingDate      *time.Time `json:"next_billing_date,omitempty"`
	CurrentPeriodEndsAt  *time.Time `json:"current_period_ends_at,omitempty"`
	PaddleSubscriptionID string     `json:"paddle_subscription_id,omitempty"`
	PaddleCustomerID     string     `json:"paddle_customer_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// SeatPeriod is a contiguous span during which seat count was constant.
// Closing a period (period_end set) and opening a new one is how mid-month
// seat changes are tracked for the Google-Workspace-style line-item invoice.
type SeatPeriod struct {
	ID                int64      `json:"id"`
	SubscriptionID    int64      `json:"subscription_id"`
	OwnerID           int64      `json:"owner_id"`
	SeatCount         int        `json:"seat_count"`
	PricePerSeatCents int        `json:"price_per_seat_cents"`
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         *time.Time `json:"period_end"`
	BillingMonth      time.Time  `json:"billing_month"`
}

// BillingLineItem is a displayable line computed from a SeatPeriod.
// Amount = billable_seat_count * days * (price_per_seat / days_in_month).
type BillingLineItem struct {
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         *time.Time `json:"period_end"` // nil = "present"
	SeatCount         int        `json:"seat_count"`
	IncludedSeatLimit int        `json:"included_seat_limit"`
	IncludedSeatCount int        `json:"included_seat_count"`
	BillableSeatCount int        `json:"billable_seat_count"`
	IncludedCents     int        `json:"included_cents"`
	PricePerSeatCents int        `json:"price_per_seat_cents"`
	AmountCents       int        `json:"amount_cents"`
	Type              string     `json:"type"`
	Description       string     `json:"description"`
	IsGranted         bool       `json:"is_granted,omitempty"`
	GrantExpiresAt    *time.Time `json:"grant_expires_at,omitempty"`
	AddonID           string     `json:"addon_id,omitempty"`
	Quantity          int        `json:"quantity"`
	UnitPriceCents    int        `json:"unit_price_cents"`
}

// AddonBillingPeriod is a paid addon span that overlaps a billing month.
type AddonBillingPeriod struct {
	AddonID        string     `json:"addon_id"`
	Name           string     `json:"name"`
	PriceCents     int        `json:"price_cents"`
	PeriodStart    time.Time  `json:"period_start"`
	PeriodEnd      *time.Time `json:"period_end"`
	BillingMonth   time.Time  `json:"billing_month"`
	IsGranted      bool       `json:"is_granted"`
	GrantExpiresAt *time.Time `json:"grant_expires_at,omitempty"`
}

// BillingDetail is the full view of the current billing month.
type BillingDetail struct {
	BillingMonth         time.Time         `json:"billing_month"`
	NextBillingDate      *time.Time        `json:"next_billing_date"`
	LineItems            []BillingLineItem `json:"line_items"`
	TotalCents           int               `json:"total_cents"`
	PlanName             string            `json:"plan_name"`
	PricePerSeatCents    int               `json:"price_per_seat_cents"`
	CurrentSeatCount     int               `json:"current_seat_count"`
	IncludedSeatLimit    int               `json:"included_seat_limit"`
	IncludedSeatCount    int               `json:"included_seat_count"`
	BillableSeatCount    int               `json:"billable_seat_count"`
	SeatDiscountCents    int               `json:"seat_discount_cents"`
	Status               string            `json:"status"`
	IsGranted            bool              `json:"is_granted"`
	GrantExpiresAt       *time.Time        `json:"grant_expires_at"`
	TrialEndsAt          *time.Time        `json:"trial_ends_at"`
	PaddleSubscriptionID string            `json:"paddle_subscription_id"`
}

// BillingHistoryMonth is a past month's invoice summary.
type BillingHistoryMonth struct {
	Month      time.Time         `json:"month"`
	TotalCents int               `json:"total_cents"`
	LineItems  []BillingLineItem `json:"line_items"`
	Status     string            `json:"status"` // "paid" | "draft"
}

type Repository interface {
	ListPlans(ctx context.Context) ([]Plan, error)
	GetPlanBySlug(ctx context.Context, slug string) (*Plan, error)
	GetPlanByID(ctx context.Context, id int64) (*Plan, error)
	GetSubscription(ctx context.Context, ownerID int64) (*Subscription, error)
	UpsertSubscription(ctx context.Context, ownerID, planID int64, trialDays, seatCount int) error
	GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error)
	ExpireTrialSubscriptions(ctx context.Context) error
	CountActiveSeats(ctx context.Context, ownerID int64) (int, error)
	// Seat period management
	GetPeriodsForMonth(ctx context.Context, ownerID int64, month time.Time) ([]SeatPeriod, error)
	GetAddonPeriodsForMonth(ctx context.Context, ownerID, businessID int64, month time.Time) ([]AddonBillingPeriod, error)
	GetActivePeriod(ctx context.Context, subscriptionID int64) (*SeatPeriod, error)
	CloseActivePeriod(ctx context.Context, subscriptionID int64, periodEnd time.Time) error
	OpenPeriod(ctx context.Context, subscriptionID, ownerID int64, seatCount, pricePerSeatCents int, periodStart, billingMonth time.Time) error
	UpdateSeatCount(ctx context.Context, ownerID int64, seatCount int) error
	GetBillingHistory(ctx context.Context, ownerID int64, limit int) ([]BillingHistoryMonth, error)
	UpdatePaddleSubscription(ctx context.Context, ownerID int64, paddleSubID, paddleCustID, status string) (bool, error)
	ApplyPaddleSubscriptionEvent(ctx context.Context, eventID, paddleSubID string, occurredAt time.Time, fallbackOwnerID int64, paddleCustID, status string) (bool, int64, error)
	RecordPaddleEvent(ctx context.Context, eventID, paddleSubID string, occurredAt time.Time) (bool, error)
	LastPaddleEventAt(ctx context.Context, paddleSubID string) (time.Time, bool, error)
	GetOwnerIDByPaddleSubscription(ctx context.Context, paddleSubID string) (int64, error)
}
