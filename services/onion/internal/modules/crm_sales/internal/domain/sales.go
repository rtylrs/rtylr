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
	ID           int64         `json:"id"`
	OwnerID      int64         `json:"owner_id"`
	BusinessID   int64         `json:"business_id"`
	CompanyName  string        `json:"company_name"`
	Industry     string        `json:"industry"`
	Size         string        `json:"size"`
	Source       string        `json:"source"`
	Stage        LeadStage     `json:"stage"`
	AssignedTo   *int64        `json:"assigned_to"`
	MRRPotential *money.Amount `json:"mrr_potential"`
	Notes        string        `json:"notes"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type CreateLeadRequest struct {
	CompanyName  string        `json:"company_name"`
	Industry     string        `json:"industry"`
	Size         string        `json:"size"`
	Source       string        `json:"source"`
	Stage        LeadStage     `json:"stage"`
	AssignedTo   *int64        `json:"assigned_to"`
	MRRPotential *money.Amount `json:"mrr_potential"`
	Notes        string        `json:"notes"`
}

type UpdateLeadRequest struct {
	CompanyName  string        `json:"company_name"`
	Industry     string        `json:"industry"`
	Size         string        `json:"size"`
	Source       string        `json:"source"`
	Stage        LeadStage     `json:"stage"`
	AssignedTo   *int64        `json:"assigned_to"`
	MRRPotential *money.Amount `json:"mrr_potential"`
	Notes        string        `json:"notes"`
}

type Contact struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	LeadID     int64     `json:"lead_id"`
	CustomerID int64     `json:"customer_id"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	LinkedIn   string    `json:"linkedin"`
	IsPrimary  bool      `json:"is_primary"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateContactRequest struct {
	LeadID    int64  `json:"lead_id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	LinkedIn  string `json:"linkedin"`
	IsPrimary bool   `json:"is_primary"`
	Notes     string `json:"notes"`
}

type UpdateContactRequest struct {
	Name      string `json:"name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	LinkedIn  string `json:"linkedin"`
	IsPrimary bool   `json:"is_primary"`
	Notes     string `json:"notes"`
}

type Touchpoint struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	LeadID       int64     `json:"lead_id"`
	ContactID    *int64    `json:"contact_id"`
	Type         string    `json:"type"`
	Outcome      string    `json:"outcome"`
	Notes        string    `json:"notes"`
	NextFollowUp *string   `json:"next_follow_up"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateTouchpointRequest struct {
	LeadID       int64   `json:"lead_id"`
	ContactID    *int64  `json:"contact_id"`
	Type         string  `json:"type"`
	Outcome      string  `json:"outcome"`
	Notes        string  `json:"notes"`
	NextFollowUp *string `json:"next_follow_up"`
}

type Contract struct {
	ID           int64        `json:"id"`
	OwnerID      int64        `json:"owner_id"`
	BusinessID   int64        `json:"business_id"`
	LeadID       int64        `json:"lead_id"`
	Plan         string       `json:"plan"`
	MRR          money.Amount `json:"mrr"`
	Seats        int          `json:"seats"`
	SignedAt     *string      `json:"signed_at"`
	SignedByName string       `json:"signed_by_name"`
	ClosedBy     *int64       `json:"closed_by"`
	Notes        string       `json:"notes"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type CreateContractRequest struct {
	LeadID       int64        `json:"lead_id"`
	Plan         string       `json:"plan"`
	MRR          money.Amount `json:"mrr"`
	Seats        int          `json:"seats"`
	SignedAt     *string      `json:"signed_at"`
	SignedByName string       `json:"signed_by_name"`
	ClosedBy     *int64       `json:"closed_by"`
	Notes        string       `json:"notes"`
}

type UpdateContractRequest struct {
	Plan         string       `json:"plan"`
	MRR          money.Amount `json:"mrr"`
	Seats        int          `json:"seats"`
	SignedAt     *string      `json:"signed_at"`
	SignedByName string       `json:"signed_by_name"`
	ClosedBy     *int64       `json:"closed_by"`
	Notes        string       `json:"notes"`
}

type PipelineSummary struct {
	Stage    string       `json:"stage"`
	Count    int          `json:"count"`
	TotalMRR money.Amount `json:"total_mrr"`
}

// LeadPage is a single page of leads plus pagination metadata.
type LeadPage struct {
	Leads  []Lead `json:"leads"`
	Total  int64  `json:"total"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

// ContactPage is a single page of contacts plus pagination metadata.
type ContactPage struct {
	Contacts []Contact `json:"contacts"`
	Total    int64     `json:"total"`
	Limit    int32     `json:"limit"`
	Offset   int32     `json:"offset"`
}

// ContractPage is a single page of contracts plus pagination metadata.
type ContractPage struct {
	Contracts []Contract `json:"contracts"`
	Total     int64      `json:"total"`
	Limit     int32      `json:"limit"`
	Offset    int32      `json:"offset"`
}
