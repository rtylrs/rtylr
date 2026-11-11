package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrRegionPriceNotFound  = errors.New("no chargeable regional price")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

const FreeIncludedSeatLimit = 10

// DefaultCurrency is what the plans table's own prices are quoted in.
const DefaultCurrency = "USD"

type Plan struct {
	ID                      int64
	Name                    string
	Slug                    string
	Apps                    []string
	PricePerSeatCents       int
	AnnualPricePerSeatCents int
	TrialDays               int
	PaddlePriceID           string
	Tagline                 string
	MaxSeats                *int
	MaxLocations            *int
	AICreditsPerMonth       int
	Features                []string
	TierOrder               int
	IsSolo                  bool
	IsFeatured              bool
	IsActive                bool
	CreatedAt               time.Time
	// Currency and Region describe the prices above. They are USD and empty
	// unless a regional price replaced them (see WithRegionPrice).
	Currency string
	Region   string
}

// RegionPrice is a plan's price override for one region.
type RegionPrice struct {
	PlanID                  int64
	Region                  string
	Currency                string
	PricePerSeatCents       int
	AnnualPricePerSeatCents int
	PaddlePriceID           string
}

// WithRegionPrice returns the plan quoted at the regional price.
func (p Plan) WithRegionPrice(rp RegionPrice) Plan {
	p.PricePerSeatCents = rp.PricePerSeatCents
	p.AnnualPricePerSeatCents = rp.AnnualPricePerSeatCents
	if rp.PaddlePriceID != "" {
		p.PaddlePriceID = rp.PaddlePriceID
	}
	p.Currency = rp.Currency
	p.Region = rp.Region
	return p
}

type Subscription struct {
	ID                   int64
	BusinessID           int64
	PlanID               int64
	PlanName             string
	PlanSlug             string
	AICreditsPerMonth    int
	Apps                 []string
	PricePerSeatCents    int
	Status               string
	IsGranted            bool
	SeatCount            int
	IncludedSeatLimit    int
	IncludedSeatCount    int
	BillableSeatCount    int
	GrantExpiresAt       *time.Time
	TrialEndsAt          *time.Time
	NextBillingDate      *time.Time
	CurrentPeriodEndsAt  *time.Time
	PaddleSubscriptionID string
	PaddleCustomerID     string
	CreatedAt            time.Time
}

// SeatPeriod is a contiguous span during which seat count was constant.
// Closing a period (period_end set) and opening a new one is how mid-month
// seat changes are tracked for the Google-Workspace-style line-item invoice.
type SeatPeriod struct {
	ID                int64
	SubscriptionID    int64
	BusinessID        int64
	SeatCount         int
	PricePerSeatCents int
	PeriodStart       time.Time
	PeriodEnd         *time.Time
	BillingMonth      time.Time
}

// BillingLineItem is a displayable line computed from a SeatPeriod.
// Amount = billable_seat_count * days * (price_per_seat / days_in_month).
type BillingLineItem struct {
	PeriodStart       time.Time
	PeriodEnd         *time.Time // nil = "present"
	SeatCount         int
	IncludedSeatLimit int
	IncludedSeatCount int
	BillableSeatCount int
	IncludedCents     int
	PricePerSeatCents int
	AmountCents       int
	Type              string
	Description       string
	IsGranted         bool
	GrantExpiresAt    *time.Time
	AddonID           string
	Quantity          int
	UnitPriceCents    int
}

// AddonBillingPeriod is a paid addon span that overlaps a billing month.
type AddonBillingPeriod struct {
	AddonID        string
	Name           string
	PriceCents     int
	PeriodStart    time.Time
	PeriodEnd      *time.Time
	BillingMonth   time.Time
	IsGranted      bool
	GrantExpiresAt *time.Time
}

// BillingDetail is the full view of the current billing month.
type BillingDetail struct {
	BillingMonth         time.Time
	NextBillingDate      *time.Time
	LineItems            []BillingLineItem
	TotalCents           int
	PlanName             string
	PricePerSeatCents    int
	CurrentSeatCount     int
	IncludedSeatLimit    int
	IncludedSeatCount    int
	BillableSeatCount    int
	SeatDiscountCents    int
	Status               string
	IsGranted            bool
	GrantExpiresAt       *time.Time
	TrialEndsAt          *time.Time
	PaddleSubscriptionID string
}

// BillingHistoryMonth is a past month's invoice summary.
type BillingHistoryMonth struct {
	Month      time.Time
	TotalCents int
	LineItems  []BillingLineItem
	Status     string // "paid" | "draft"
}

type Repository interface {
	ListPlans(ctx context.Context) ([]Plan, error)
	ListRegionPrices(ctx context.Context, region string) ([]RegionPrice, error)
	GetRegionPrice(ctx context.Context, planID int64, region string) (*RegionPrice, error)
	GetPlanBySlug(ctx context.Context, slug string) (*Plan, error)
	GetPlanByID(ctx context.Context, id int64) (*Plan, error)
	GetSubscription(ctx context.Context, businessID int64) (*Subscription, error)
	UpsertSubscription(ctx context.Context, businessID, planID int64, trialDays, seatCount int) error
	GetBusinessApps(ctx context.Context, businessID int64) ([]string, error)
	ExpireTrialSubscriptions(ctx context.Context) error
	CountActiveSeats(ctx context.Context, businessID int64) (int, error)
	// Seat period management
	GetPeriodsForMonth(ctx context.Context, businessID int64, month time.Time) ([]SeatPeriod, error)
	GetAddonPeriodsForMonth(ctx context.Context, businessID int64, month time.Time) ([]AddonBillingPeriod, error)
	// ReconcileSeatBilling is the only writer of seat_periods and
	// subscriptions.seat_count. It creates the subscription row for a business
	// that has none, using defaultPlanID/defaultTrialDays, and then rolls the
	// seat ledger forward to asOf. It takes a row lock on the subscription, so
	// concurrent callers serialise rather than each opening a period.
	ReconcileSeatBilling(ctx context.Context, businessID, defaultPlanID int64, defaultTrialDays, seatCount int, asOf time.Time) error
	// ListBusinessIDsForReconciliation pages the business ids that may need a
	// seat ledger, ascending and strictly greater than afterBusinessID.
	ListBusinessIDsForReconciliation(ctx context.Context, afterBusinessID int64, limit int) ([]int64, error)
	GetBillingHistory(ctx context.Context, businessID int64, limit int) ([]BillingHistoryMonth, error)
	UpdatePaddleSubscription(ctx context.Context, businessID int64, paddleSubID, paddleCustID, status string) (bool, error)
	ApplyPaddleSubscriptionEvent(ctx context.Context, eventID, paddleSubID string, occurredAt time.Time, fallbackBusinessID int64, paddleCustID, status string) (bool, int64, error)
	RecordPaddleEvent(ctx context.Context, eventID, paddleSubID string, occurredAt time.Time) (bool, error)
	LastPaddleEventAt(ctx context.Context, paddleSubID string) (time.Time, bool, error)
	GetBusinessIDByPaddleSubscription(ctx context.Context, paddleSubID string) (int64, error)
}
