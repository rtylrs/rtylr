package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

// Billing is how often a value is charged: by a deal, a lead's estimate or a
// contract.
type Billing string

const (
	BillingOneTime Billing = "one_time"
	BillingMonthly Billing = "monthly"
	BillingYearly  Billing = "yearly"
)

func (b Billing) IsValid() bool {
	return b == BillingOneTime || b == BillingMonthly || b == BillingYearly
}

// MonthlyEquivalent is recurring revenue per month: the value itself when
// billed monthly, a twelfth of it when yearly, and nothing for a one-time
// sale. Contracts keep it as MRR for readers that predate value and billing.
func MonthlyEquivalent(value money.Amount, billing Billing) money.Amount {
	switch billing {
	case BillingMonthly:
		return value
	case BillingYearly:
		return value / 12
	default:
		return 0
	}
}

// FirstYearAmount is what a sale brings in over its first year, the common
// unit pipeline totals add up: twelve months of a monthly price, the value of
// a yearly or one-time sale.
func FirstYearAmount(value money.Amount, billing Billing) money.Amount {
	if billing == BillingMonthly {
		return value * 12
	}
	return value
}

// Contract records what a won deal sold. Plan is free text for what was sold;
// Value is charged once or per Billing period over the optional term from
// StartDate to EndDate (YYYY-MM-DD). A contract is billable once SignedAt is
// set.
type Contract struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	LeadID       *int64
	DealID       *int64
	Plan         string
	MRR          money.Amount
	Seats        int
	Value        money.Amount
	Billing      Billing
	StartDate    *string
	EndDate      *string
	CustomerID   *int64
	SignedAt     *string
	SignedByName string
	ClosedBy     *int64
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ContractInput is a contract's editable terms.
type ContractInput struct {
	DealID       *int64
	Plan         string
	Seats        int
	Value        money.Amount
	Billing      Billing
	StartDate    *string
	EndDate      *string
	CustomerID   *int64
	SignedAt     *string
	SignedByName string
	Notes        string
}

// Normalize fills defaults and checks the terms.
func (in *ContractInput) Normalize() error {
	in.Plan = trimTo(in.Plan, 255)
	in.SignedByName = trimTo(in.SignedByName, 255)
	if in.Plan == "" {
		return invalid("plan is required")
	}
	if in.Seats <= 0 {
		in.Seats = 1
	}
	if in.Billing == "" {
		in.Billing = BillingMonthly
	}
	if !in.Billing.IsValid() {
		return invalid("billing must be one_time, monthly or yearly")
	}
	if in.Value < 0 {
		return invalid("value cannot be negative")
	}
	for _, date := range []*string{in.StartDate, in.EndDate, in.SignedAt} {
		if date != nil && *date != "" && !IsDate(*date) {
			return invalid("dates must be YYYY-MM-DD")
		}
	}
	if in.StartDate != nil && in.EndDate != nil && *in.StartDate != "" && *in.EndDate != "" && *in.EndDate < *in.StartDate {
		return invalid("the contract cannot end before it starts")
	}
	return nil
}

// ContractPage is a single page of contracts plus pagination metadata.
type ContractPage struct {
	Contracts []Contract
	Total     int64
	Limit     int32
	Offset    int32
}

// Rep is a teammate work can be assigned to: the business owner or an active
// member.
type Rep struct {
	ID   int64
	Name string
}
