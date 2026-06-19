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
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(crmSalesScope, businessID))
	}
}

type repository interface {
	ListLeads(ctx context.Context, businessID int64) ([]domain.Lead, error)
	GetLead(ctx context.Context, id, businessID int64) (domain.Lead, error)
	CreateLead(ctx context.Context, businessID int64, req domain.CreateLeadRequest) (domain.Lead, error)
	UpdateLead(ctx context.Context, id, businessID int64, req domain.UpdateLeadRequest) (domain.Lead, error)
	DeleteLead(ctx context.Context, id, businessID int64) error
	GetPipeline(ctx context.Context, businessID int64) ([]domain.PipelineSummary, error)

	ListContacts(ctx context.Context, businessID int64, leadID *int64) ([]domain.Contact, error)
	CreateContact(ctx context.Context, businessID int64, req domain.CreateContactRequest) (domain.Contact, error)
	UpdateContact(ctx context.Context, id, businessID int64, req domain.UpdateContactRequest) (domain.Contact, error)
	DeleteContact(ctx context.Context, id, businessID int64) error

	ListTouchpoints(ctx context.Context, businessID, leadID int64) ([]domain.Touchpoint, error)
	CreateTouchpoint(ctx context.Context, businessID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error)
	DeleteTouchpoint(ctx context.Context, id, businessID int64) error

	ListContracts(ctx context.Context, businessID int64) ([]domain.Contract, error)
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

func (s *Service) ListLeads(ctx context.Context, businessID int64) ([]domain.Lead, error) {
	return cacher.Of[[]domain.Lead]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(crmSalesScope, businessID, "leads"),
		func(ctx context.Context) ([]domain.Lead, error) {
			return s.repo.ListLeads(ctx, businessID)
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
	lead, err := s.repo.CreateLead(ctx, businessID, req)
	if err == nil {
		invalidateCRMSalesScope(ctx, businessID)
	}
	return lead, err
}

func (s *Service) UpdateLead(ctx context.Context, id, businessID int64, req domain.UpdateLeadRequest) (domain.Lead, error) {
	if req.CompanyName == "" {
		req.CompanyName = "Unknown"
	}
	if req.Source == "" {
		req.Source = "cold_email"
	}
	if req.Stage == "" {
		req.Stage = domain.LeadStageCold
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

func (s *Service) ListContacts(ctx context.Context, businessID int64, leadID *int64) ([]domain.Contact, error) {
	lead := int64(0)
	if leadID != nil {
		lead = *leadID
	}
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contacts|l=%d", lead))
	return cacher.Of[[]domain.Contact]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Contact, error) {
			return s.repo.ListContacts(ctx, businessID, leadID)
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

func (s *Service) ListContracts(ctx context.Context, businessID int64) ([]domain.Contract, error) {
	return cacher.Of[[]domain.Contract]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(crmSalesScope, businessID, "contracts"),
		func(ctx context.Context) ([]domain.Contract, error) {
			return s.repo.ListContracts(ctx, businessID)
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
