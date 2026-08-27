package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	crmSalesScope = "crm_sales"
	crmSalesTTL   = 5 * time.Minute
)

func invalidateCRMSalesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(crmSalesScope, businessID))
}

type repository interface {
	ListLeads(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Lead, error)
	CountLeads(ctx context.Context, businessID int64) (int64, error)
	GetLead(ctx context.Context, id, businessID int64) (domain.Lead, error)
	CreateLead(ctx context.Context, businessID int64, req domain.CreateLeadRequest) (domain.Lead, error)
	UpdateLead(ctx context.Context, id, businessID int64, req domain.UpdateLeadRequest) (domain.Lead, error)
	ValidateLeadAssignee(ctx context.Context, businessID, userID int64) error
	DeleteLead(ctx context.Context, id, businessID int64) error
	GetPipeline(ctx context.Context, businessID int64) ([]domain.PipelineSummary, error)
	ListSignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, error)
	CountSignedContracts(ctx context.Context, businessID int64) (int64, error)
	ListUnsignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, error)
	CountUnsignedContracts(ctx context.Context, businessID int64) (int64, error)

	ListContacts(ctx context.Context, businessID int64, leadID *int64, limit, offset int32) ([]domain.Contact, error)
	CountContacts(ctx context.Context, businessID int64, leadID *int64) (int64, error)
	CreateContact(ctx context.Context, businessID int64, req domain.CreateContactRequest) (domain.Contact, error)
	UpdateContact(ctx context.Context, id, businessID int64, req domain.UpdateContactRequest) (domain.Contact, error)
	DeleteContact(ctx context.Context, id, businessID int64) error

	ListTouchpoints(ctx context.Context, businessID, leadID int64) ([]domain.Touchpoint, error)
	CreateTouchpoint(ctx context.Context, businessID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error)
	DeleteTouchpoint(ctx context.Context, id, businessID int64) error

	ListContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, error)
	CountContracts(ctx context.Context, businessID int64) (int64, error)
	GetContract(ctx context.Context, id, businessID int64) (domain.Contract, error)
	CreateContract(ctx context.Context, businessID int64, req domain.CreateContractRequest) (domain.Contract, error)
	UpdateContract(ctx context.Context, id, businessID int64, req domain.UpdateContractRequest) (domain.Contract, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// normalizePagination applies the standard list defaults: limit falls back to
// 20 when unset, is capped at 100, and offset is clamped to zero or above.
func normalizePagination(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (s *Service) ListLeads(ctx context.Context, businessID int64, limit, offset int32) (domain.LeadPage, error) {
	limit, offset = normalizePagination(limit, offset)
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("leads|n=%d|o=%d", limit, offset))
	return cacher.Of[domain.LeadPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.LeadPage, error) {
			leads, err := s.repo.ListLeads(ctx, businessID, limit, offset)
			if err != nil {
				return domain.LeadPage{}, err
			}
			total, err := s.repo.CountLeads(ctx, businessID)
			if err != nil {
				return domain.LeadPage{}, err
			}
			return domain.LeadPage{Leads: leads, Total: total, Limit: limit, Offset: offset}, nil
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) GetLead(ctx context.Context, id, businessID int64) (domain.Lead, error) {
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("lead=%d", id))
	return cacher.Of[domain.Lead]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Lead, error) {
			return s.repo.GetLead(ctx, id, businessID)
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) CreateLead(ctx context.Context, businessID int64, req domain.CreateLeadRequest) (domain.Lead, error) {
	if req.CompanyName == "" {
		req.CompanyName = "Unknown"
	}
	if req.Source == "" {
		req.Source = "cold_email"
	}
	if req.Stage == "" {
		req.Stage = domain.LeadStageCold
	}
	if !req.Stage.IsValid() {
		return domain.Lead{}, fmt.Errorf("%w: %q", domain.ErrInvalidLeadStage, string(req.Stage))
	}
	if req.AssignedTo != nil {
		if err := s.repo.ValidateLeadAssignee(ctx, businessID, *req.AssignedTo); err != nil {
			return domain.Lead{}, err
		}
	}
	lead, err := s.repo.CreateLead(ctx, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return lead, err
}

func (s *Service) UpdateLead(ctx context.Context, id, businessID int64, req domain.UpdateLeadRequest) (domain.Lead, error) {
	current, err := s.repo.GetLead(ctx, id, businessID)
	if err != nil {
		return domain.Lead{}, err
	}
	if req.CompanyName == "" {
		req.CompanyName = current.CompanyName
	}
	if req.Source == "" {
		req.Source = current.Source
	}
	if req.Stage == "" {
		req.Stage = current.Stage
	}
	if !req.Stage.IsValid() {
		return domain.Lead{}, fmt.Errorf("%w: %q", domain.ErrInvalidLeadStage, string(req.Stage))
	}
	if !current.Stage.CanTransitionTo(req.Stage) {
		return domain.Lead{}, fmt.Errorf("%w: %s -> %s", domain.ErrInvalidLeadTransition, current.Stage, req.Stage)
	}
	if req.AssignedTo != nil {
		if err := s.repo.ValidateLeadAssignee(ctx, businessID, *req.AssignedTo); err != nil {
			return domain.Lead{}, err
		}
	}
	lead, err := s.repo.UpdateLead(ctx, id, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return lead, err
}

func (s *Service) DeleteLead(ctx context.Context, id, businessID int64) error {
	if err := s.repo.DeleteLead(ctx, id, businessID); err != nil {
		return err
	}
	invalidateCRMSalesScope(ctx, businessID)
	return nil
}

func (s *Service) GetPipeline(ctx context.Context, businessID int64) ([]domain.PipelineSummary, error) {
	return cacher.Of[[]domain.PipelineSummary]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(crmSalesScope, businessID, "pipeline"),
		func(ctx context.Context) ([]domain.PipelineSummary, error) {
			return s.repo.GetPipeline(ctx, businessID)
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) ListContacts(ctx context.Context, businessID int64, leadID *int64, limit, offset int32) (domain.ContactPage, error) {
	limit, offset = normalizePagination(limit, offset)
	lead := int64(0)
	if leadID != nil {
		lead = *leadID
	}
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contacts|l=%d|n=%d|o=%d", lead, limit, offset))
	return cacher.Of[domain.ContactPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.ContactPage, error) {
			contacts, err := s.repo.ListContacts(ctx, businessID, leadID, limit, offset)
			if err != nil {
				return domain.ContactPage{}, err
			}
			total, err := s.repo.CountContacts(ctx, businessID, leadID)
			if err != nil {
				return domain.ContactPage{}, err
			}
			return domain.ContactPage{Contacts: contacts, Total: total, Limit: limit, Offset: offset}, nil
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) CreateContact(ctx context.Context, businessID int64, req domain.CreateContactRequest) (domain.Contact, error) {
	c, err := s.repo.CreateContact(ctx, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return c, err
}

func (s *Service) UpdateContact(ctx context.Context, id, businessID int64, req domain.UpdateContactRequest) (domain.Contact, error) {
	c, err := s.repo.UpdateContact(ctx, id, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return c, err
}

func (s *Service) DeleteContact(ctx context.Context, id, businessID int64) error {
	if err := s.repo.DeleteContact(ctx, id, businessID); err != nil {
		return err
	}
	invalidateCRMSalesScope(ctx, businessID)
	return nil
}

func (s *Service) ListTouchpoints(ctx context.Context, businessID, leadID int64) ([]domain.Touchpoint, error) {
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("touchpoints|l=%d", leadID))
	return cacher.Of[[]domain.Touchpoint]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Touchpoint, error) {
			return s.repo.ListTouchpoints(ctx, businessID, leadID)
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) CreateTouchpoint(ctx context.Context, businessID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error) {
	if req.Type == "" {
		req.Type = "call"
	}
	if req.Outcome == "" {
		req.Outcome = "other"
	}
	tp, err := s.repo.CreateTouchpoint(ctx, businessID, createdBy, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return tp, err
}

func (s *Service) DeleteTouchpoint(ctx context.Context, id, businessID int64) error {
	if err := s.repo.DeleteTouchpoint(ctx, id, businessID); err != nil {
		return err
	}
	invalidateCRMSalesScope(ctx, businessID)
	return nil
}

func (s *Service) ListContracts(ctx context.Context, businessID int64, limit, offset int32) (domain.ContractPage, error) {
	limit, offset = normalizePagination(limit, offset)
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contracts|n=%d|o=%d", limit, offset))
	return cacher.Of[domain.ContractPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.ContractPage, error) {
			contracts, err := s.repo.ListContracts(ctx, businessID, limit, offset)
			if err != nil {
				return domain.ContractPage{}, err
			}
			total, err := s.repo.CountContracts(ctx, businessID)
			if err != nil {
				return domain.ContractPage{}, err
			}
			return domain.ContractPage{Contracts: contracts, Total: total, Limit: limit, Offset: offset}, nil
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) GetContract(ctx context.Context, id, businessID int64) (domain.Contract, error) {
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contract=%d", id))
	return cacher.Of[domain.Contract]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Contract, error) {
			return s.repo.GetContract(ctx, id, businessID)
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) CreateContract(ctx context.Context, businessID int64, req domain.CreateContractRequest) (domain.Contract, error) {
	if req.Plan == "" {
		req.Plan = "pro"
	}
	if req.Seats == 0 {
		req.Seats = 1
	}
	c, err := s.repo.CreateContract(ctx, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return c, err
}

func (s *Service) UpdateContract(ctx context.Context, id, businessID int64, req domain.UpdateContractRequest) (domain.Contract, error) {
	if req.Plan == "" {
		req.Plan = "pro"
	}
	if req.Seats == 0 {
		req.Seats = 1
	}
	c, err := s.repo.UpdateContract(ctx, id, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return c, err
}

// ListSignedContracts returns the page of contracts whose signature is
// complete. Backs crm.sales.v1's signed=true filter (#839): a caller asking
// for billable contracts must not page through drafts, and the total must be
// the filtered count rather than a full count trimmed afterwards.
func (s *Service) ListSignedContracts(ctx context.Context, businessID int64, limit, offset int32) (domain.ContractPage, error) {
	return s.listContractsBySignState(ctx, businessID, true, limit, offset)
}

// ListUnsignedContracts is the signed=false variant of the above.
func (s *Service) ListUnsignedContracts(ctx context.Context, businessID int64, limit, offset int32) (domain.ContractPage, error) {
	return s.listContractsBySignState(ctx, businessID, false, limit, offset)
}

func (s *Service) listContractsBySignState(ctx context.Context, businessID int64, signed bool, limit, offset int32) (domain.ContractPage, error) {
	limit, offset = normalizePagination(limit, offset)
	state := "unsigned"
	if signed {
		state = "signed"
	}
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contracts|%s|n=%d|o=%d", state, limit, offset))
	return cacher.Of[domain.ContractPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.ContractPage, error) {
			var (
				contracts []domain.Contract
				total     int64
				err       error
			)
			if signed {
				contracts, err = s.repo.ListSignedContracts(ctx, businessID, limit, offset)
				if err == nil {
					total, err = s.repo.CountSignedContracts(ctx, businessID)
				}
			} else {
				contracts, err = s.repo.ListUnsignedContracts(ctx, businessID, limit, offset)
				if err == nil {
					total, err = s.repo.CountUnsignedContracts(ctx, businessID)
				}
			}
			if err != nil {
				return domain.ContractPage{}, err
			}
			return domain.ContractPage{Contracts: contracts, Total: total, Limit: limit, Offset: offset}, nil
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}
