package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

// DealStatus is whether a deal is still being worked or how it closed.
type DealStatus string

const (
	DealOpen DealStatus = "open"
	DealWon  DealStatus = "won"
	DealLost DealStatus = "lost"
)

func (s DealStatus) IsValid() bool {
	return s == DealOpen || s == DealWon || s == DealLost
}

// StatusForStage is the status a deal takes in a stage of a kind.
func StatusForStage(kind StageKind) DealStatus {
	switch kind {
	case StageWon:
		return DealWon
	case StageLost:
		return DealLost
	}
	return DealOpen
}

// Deal is one opportunity in a pipeline.
type Deal struct {
	ID                int64
	BusinessID        int64
	OwnerID           int64
	PipelineID        int64
	StageID           int64
	Title             string
	LeadID            *int64
	CustomerID        *int64
	CompanyID         *int64
	ContactName       string
	CompanyName       string
	Value             money.Amount
	Billing           Billing
	ExpectedCloseDate string
	AssignedTo        *int64
	Source            string
	Status            DealStatus
	LostReason        string
	ClosedAt          *time.Time
	Notes             string
	LastActivityAt    *time.Time
	NextFollowUp      string
	StageChangedAt    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Amount is the deal's first-year value, the unit pipeline totals use.
func (d Deal) Amount() money.Amount {
	return FirstYearAmount(d.Value, d.Billing)
}

// DealInput is a deal's editable fields. PipelineID and StageID only apply on
// creation; moves go through MoveDeal.
type DealInput struct {
	Title             string
	PipelineID        int64
	StageID           int64
	ContactName       string
	CompanyName       string
	Value             money.Amount
	Billing           Billing
	ExpectedCloseDate string
	AssignedTo        *int64
	Source            string
	Notes             string
}

func (in *DealInput) Normalize() error {
	in.Title = trimTo(in.Title, 255)
	in.ContactName = trimTo(in.ContactName, 255)
	in.CompanyName = trimTo(in.CompanyName, 255)
	in.Source = trimTo(in.Source, 100)
	in.Notes = trimTo(in.Notes, 10000)
	if in.Title == "" {
		in.Title = in.CompanyName
	}
	if in.Title == "" {
		in.Title = in.ContactName
	}
	if in.Title == "" {
		return invalid("a deal needs a title")
	}
	if in.Billing == "" {
		in.Billing = BillingOneTime
	}
	if !in.Billing.IsValid() {
		return invalid("billing must be one_time, monthly or yearly")
	}
	if in.Value < 0 {
		return invalid("value cannot be negative")
	}
	if in.ExpectedCloseDate != "" && !IsDate(in.ExpectedCloseDate) {
		return invalid("expected close date must be YYYY-MM-DD")
	}
	if in.AssignedTo != nil && *in.AssignedTo <= 0 {
		in.AssignedTo = nil
	}
	return nil
}

// DealSort orders a deal list.
type DealSort string

const (
	DealSortNewest    DealSort = "newest"
	DealSortCloseDate DealSort = "close_date"
	DealSortValue     DealSort = "value"
)

func (s DealSort) IsValid() bool {
	return s == DealSortNewest || s == DealSortCloseDate || s == DealSortValue
}

// DealFilter selects deals. AssignedTo 0 means unassigned; empty Statuses
// means every status.
type DealFilter struct {
	PipelineID *int64
	StageID    *int64
	Statuses   []DealStatus
	AssignedTo *int64
	Search     string
	Sort       DealSort
	Cursor     string
	Limit      int
}

// DealPage is one page of a deal list.
type DealPage struct {
	Deals      []Deal
	NextCursor string
	Total      int64
}

// StageTotal is a board column's header: how many deals and what they are
// worth, raw and weighted by the stage's probability.
type StageTotal struct {
	Stage          Stage
	Deals          int64
	Amount         money.Amount
	WeightedAmount money.Amount
}

// DealStats summarise a pipeline for its header: what is open now, what was
// won this month, and how deals closed over the last 90 days.
type DealStats struct {
	OpenDeals          int64
	OpenAmount         money.Amount
	WeightedAmount     money.Amount
	WonThisMonth       int64
	WonAmountThisMonth money.Amount
	WinRate            float64
	AverageDaysToWin   float64
	AverageWonAmount   money.Amount
}

// DealOutcome is how deals closed over a period.
type DealOutcome struct {
	Won              int64
	Lost             int64
	WonAmount        money.Amount
	AverageDaysToWin float64
}

// WinDealInput closes a deal as won, optionally drafting its contract.
type WinDealInput struct {
	Contract *ContractInput
}
