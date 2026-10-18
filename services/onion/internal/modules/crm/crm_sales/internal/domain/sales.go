package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

var (
	ErrLeadNotFound          = errors.New("lead not found")
	ErrContactNotFound       = errors.New("contact not found")
	ErrTouchpointNotFound    = errors.New("touchpoint not found")
	ErrContractNotFound      = errors.New("contract not found")
	ErrInvalidLeadStage      = errors.New("invalid lead stage")
	ErrInvalidLeadTransition = errors.New("invalid lead stage transition")
	ErrInvalidLeadAssignee   = errors.New("invalid lead assignee")
)

type LeadStage string

const (
	LeadStageCold         LeadStage = "cold"
	LeadStageContacted    LeadStage = "contacted"
	LeadStageDemoBooked   LeadStage = "demo_booked"
	LeadStageDemoDone     LeadStage = "demo_done"
	LeadStageProposalSent LeadStage = "proposal_sent"
	LeadStageContracted   LeadStage = "contracted"
	LeadStageLost         LeadStage = "lost"
)

// IsValid reports whether the stage is one of the known pipeline stages.
// Unvalidated stages used to be persisted verbatim and then surfaced as their
// own bucket in GetPipeline (GROUP BY stage), corrupting funnel reporting.
func (s LeadStage) IsValid() bool {
	switch s {
	case LeadStageCold, LeadStageContacted, LeadStageDemoBooked, LeadStageDemoDone,
		LeadStageProposalSent, LeadStageContracted, LeadStageLost:
		return true
	default:
		return false
	}
}

// CanTransitionTo enforces a forward-only pipeline and immutable terminal
// stages. A lead may be marked lost from any active stage, but cannot be
// resurrected or moved backwards after it is contracted or lost.
func (s LeadStage) CanTransitionTo(next LeadStage) bool {
	if !s.IsValid() || !next.IsValid() {
		return false
	}
	if s == next {
		return true
	}
	if s == LeadStageContracted || s == LeadStageLost {
		return false
	}
	if next == LeadStageLost {
		return true
	}
	stages := AllStages()
	var currentIndex, nextIndex int
	for i, stage := range stages {
		if stage == s {
			currentIndex = i
		}
		if stage == next {
			nextIndex = i
		}
	}
	return nextIndex >= currentIndex
}

type Lead struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	CompanyName  string
	Industry     string
	Size         string
	Source       string
	Stage        LeadStage
	AssignedTo   *int64
	MRRPotential *money.Amount
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// NextFollowUp is the follow-up date (YYYY-MM-DD) set on the lead's
	// latest touchpoint, or "" when that touchpoint set none. Logging a
	// touchpoint replaces it, so the date always names the next step.
	NextFollowUp string
}

// Rep is a teammate a lead can be assigned to: the business owner or an
// active member.
type Rep struct {
	ID   int64
	Name string
}

type CreateLeadRequest struct {
	CompanyName  string
	Industry     string
	Size         string
	Source       string
	Stage        LeadStage
	AssignedTo   *int64
	MRRPotential *money.Amount
	Notes        string
}

type UpdateLeadRequest struct {
	CompanyName  string
	Industry     string
	Size         string
	Source       string
	Stage        LeadStage
	AssignedTo   *int64
	MRRPotential *money.Amount
	Notes        string
}

type Contact struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	LeadID     int64
	CustomerID int64
	Name       string
	Role       string
	Email      string
	Phone      string
	LinkedIn   string
	IsPrimary  bool
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateContactRequest struct {
	LeadID    int64
	Name      string
	Role      string
	Email     string
	Phone     string
	LinkedIn  string
	IsPrimary bool
	Notes     string
}

type UpdateContactRequest struct {
	Name      string
	Role      string
	Email     string
	Phone     string
	LinkedIn  string
	IsPrimary bool
	Notes     string
}

type Touchpoint struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	LeadID       int64
	ContactID    *int64
	Type         string
	Outcome      string
	Notes        string
	NextFollowUp *string
	CreatedBy    int64
	CreatedAt    time.Time
}

type CreateTouchpointRequest struct {
	LeadID       int64
	ContactID    *int64
	Type         string
	Outcome      string
	Notes        string
	NextFollowUp *string
}

type Contract struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	LeadID       int64
	Plan         string
	MRR          money.Amount
	Seats        int
	SignedAt     *string
	SignedByName string
	ClosedBy     *int64
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateContractRequest struct {
	LeadID       int64
	Plan         string
	MRR          money.Amount
	Seats        int
	SignedAt     *string
	SignedByName string
	ClosedBy     *int64
	Notes        string
}

type UpdateContractRequest struct {
	Plan         string
	MRR          money.Amount
	Seats        int
	SignedAt     *string
	SignedByName string
	ClosedBy     *int64
	Notes        string
}

type PipelineSummary struct {
	Stage    string
	Count    int
	TotalMRR money.Amount
}

// LeadPage is a single page of leads plus pagination metadata.
type LeadPage struct {
	Leads  []Lead
	Total  int64
	Limit  int32
	Offset int32
}

// ContactPage is a single page of contacts plus pagination metadata.
type ContactPage struct {
	Contacts []Contact
	Total    int64
	Limit    int32
	Offset   int32
}

// ContractPage is a single page of contracts plus pagination metadata.
type ContractPage struct {
	Contracts []Contract
	Total     int64
	Limit     int32
	Offset    int32
}
