package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

type leadPayload struct {
	ID           int64            `json:"id"`
	OwnerID      int64            `json:"owner_id"`
	BusinessID   int64            `json:"business_id"`
	CompanyName  string           `json:"company_name"`
	Industry     string           `json:"industry"`
	Size         string           `json:"size"`
	Source       string           `json:"source"`
	Stage        domain.LeadStage `json:"stage"`
	AssignedTo   *int64           `json:"assigned_to"`
	MRRPotential *money.Amount    `json:"mrr_potential"`
	Notes        string           `json:"notes"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	NextFollowUp string           `json:"next_follow_up,omitempty"`
}

func payloadLead(value domain.Lead) leadPayload { return leadPayload(value) }

func payloadLeadPointer(value *domain.Lead) *leadPayload {
	if value == nil {
		return nil
	}
	result := payloadLead(*value)
	return &result
}

func payloadLeadList(values []domain.Lead) []leadPayload {
	if values == nil {
		return nil
	}
	result := make([]leadPayload, len(values))
	for i, value := range values {
		result[i] = payloadLead(value)
	}
	return result
}

type contactPayload struct {
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

func payloadContact(value domain.Contact) contactPayload { return contactPayload(value) }

func payloadContactPointer(value *domain.Contact) *contactPayload {
	if value == nil {
		return nil
	}
	result := payloadContact(*value)
	return &result
}

func payloadContactList(values []domain.Contact) []contactPayload {
	if values == nil {
		return nil
	}
	result := make([]contactPayload, len(values))
	for i, value := range values {
		result[i] = payloadContact(value)
	}
	return result
}

type touchpointPayload struct {
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

func payloadTouchpoint(value domain.Touchpoint) touchpointPayload { return touchpointPayload(value) }

func payloadTouchpointPointer(value *domain.Touchpoint) *touchpointPayload {
	if value == nil {
		return nil
	}
	result := payloadTouchpoint(*value)
	return &result
}

func payloadTouchpointList(values []domain.Touchpoint) []touchpointPayload {
	if values == nil {
		return nil
	}
	result := make([]touchpointPayload, len(values))
	for i, value := range values {
		result[i] = payloadTouchpoint(value)
	}
	return result
}

type contractPayload struct {
	ID           int64          `json:"id"`
	OwnerID      int64          `json:"owner_id"`
	BusinessID   int64          `json:"business_id"`
	LeadID       int64          `json:"lead_id"`
	Plan         string         `json:"plan"`
	MRR          money.Amount   `json:"mrr"`
	Seats        int            `json:"seats"`
	Value        money.Amount   `json:"value"`
	Billing      domain.Billing `json:"billing"`
	StartDate    *string        `json:"start_date"`
	EndDate      *string        `json:"end_date"`
	CustomerID   *int64         `json:"customer_id"`
	SignedAt     *string        `json:"signed_at"`
	SignedByName string         `json:"signed_by_name"`
	ClosedBy     *int64         `json:"closed_by"`
	Notes        string         `json:"notes"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func payloadContract(value domain.Contract) contractPayload { return contractPayload(value) }

func payloadContractPointer(value *domain.Contract) *contractPayload {
	if value == nil {
		return nil
	}
	result := payloadContract(*value)
	return &result
}

func payloadContractList(values []domain.Contract) []contractPayload {
	if values == nil {
		return nil
	}
	result := make([]contractPayload, len(values))
	for i, value := range values {
		result[i] = payloadContract(value)
	}
	return result
}

type pipelineSummaryPayload struct {
	Stage    string       `json:"stage"`
	Count    int          `json:"count"`
	TotalMRR money.Amount `json:"total_mrr"`
}

func payloadPipelineSummary(value domain.PipelineSummary) pipelineSummaryPayload {
	return pipelineSummaryPayload(value)
}

func payloadPipelineSummaryPointer(value *domain.PipelineSummary) *pipelineSummaryPayload {
	if value == nil {
		return nil
	}
	result := payloadPipelineSummary(*value)
	return &result
}

func payloadPipelineSummaryList(values []domain.PipelineSummary) []pipelineSummaryPayload {
	if values == nil {
		return nil
	}
	result := make([]pipelineSummaryPayload, len(values))
	for i, value := range values {
		result[i] = payloadPipelineSummary(value)
	}
	return result
}
