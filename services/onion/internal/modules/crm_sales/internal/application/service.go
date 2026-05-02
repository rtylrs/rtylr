package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
)

type repository interface {
	ListLeads(ctx context.Context, ownerID int64) ([]domain.Lead, error)
	GetLead(ctx context.Context, id, ownerID int64) (domain.Lead, error)
	CreateLead(ctx context.Context, ownerID int64, req domain.CreateLeadRequest) (domain.Lead, error)
	UpdateLead(ctx context.Context, id, ownerID int64, req domain.UpdateLeadRequest) (domain.Lead, error)
	DeleteLead(ctx context.Context, id, ownerID int64) error
	GetPipeline(ctx context.Context, ownerID int64) ([]domain.PipelineSummary, error)

	ListContacts(ctx context.Context, ownerID int64, leadID *int64) ([]domain.Contact, error)
	CreateContact(ctx context.Context, ownerID int64, req domain.CreateContactRequest) (domain.Contact, error)
	UpdateContact(ctx context.Context, id, ownerID int64, req domain.UpdateContactRequest) (domain.Contact, error)
	DeleteContact(ctx context.Context, id, ownerID int64) error

	ListTouchpoints(ctx context.Context, ownerID, leadID int64) ([]domain.Touchpoint, error)
	CreateTouchpoint(ctx context.Context, ownerID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error)
	DeleteTouchpoint(ctx context.Context, id, ownerID int64) error

	ListContracts(ctx context.Context, ownerID int64) ([]domain.Contract, error)
	GetContract(ctx context.Context, id, ownerID int64) (domain.Contract, error)
	CreateContract(ctx context.Context, ownerID int64, req domain.CreateContractRequest) (domain.Contract, error)
	UpdateContract(ctx context.Context, id, ownerID int64, req domain.UpdateContractRequest) (domain.Contract, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListLeads(ctx context.Context, ownerID int64) ([]domain.Lead, error) {
	return s.repo.ListLeads(ctx, ownerID)
}

func (s *Service) GetLead(ctx context.Context, id, ownerID int64) (domain.Lead, error) {
	return s.repo.GetLead(ctx, id, ownerID)
}

func (s *Service) CreateLead(ctx context.Context, ownerID int64, req domain.CreateLeadRequest) (domain.Lead, error) {
	if req.CompanyName == "" {
		req.CompanyName = "Unknown"
	}
	if req.Source == "" {
		req.Source = "cold_email"
	}
	if req.Stage == "" {
		req.Stage = domain.LeadStageCold
	}
	return s.repo.CreateLead(ctx, ownerID, req)
}

func (s *Service) UpdateLead(ctx context.Context, id, ownerID int64, req domain.UpdateLeadRequest) (domain.Lead, error) {
	if req.CompanyName == "" {
		req.CompanyName = "Unknown"
	}
	if req.Source == "" {
		req.Source = "cold_email"
	}
	if req.Stage == "" {
		req.Stage = domain.LeadStageCold
	}
	return s.repo.UpdateLead(ctx, id, ownerID, req)
}

func (s *Service) DeleteLead(ctx context.Context, id, ownerID int64) error {
	return s.repo.DeleteLead(ctx, id, ownerID)
}

func (s *Service) GetPipeline(ctx context.Context, ownerID int64) ([]domain.PipelineSummary, error) {
	return s.repo.GetPipeline(ctx, ownerID)
}

func (s *Service) ListContacts(ctx context.Context, ownerID int64, leadID *int64) ([]domain.Contact, error) {
	return s.repo.ListContacts(ctx, ownerID, leadID)
}

func (s *Service) CreateContact(ctx context.Context, ownerID int64, req domain.CreateContactRequest) (domain.Contact, error) {
	return s.repo.CreateContact(ctx, ownerID, req)
}

func (s *Service) UpdateContact(ctx context.Context, id, ownerID int64, req domain.UpdateContactRequest) (domain.Contact, error) {
	return s.repo.UpdateContact(ctx, id, ownerID, req)
}

func (s *Service) DeleteContact(ctx context.Context, id, ownerID int64) error {
	return s.repo.DeleteContact(ctx, id, ownerID)
}

func (s *Service) ListTouchpoints(ctx context.Context, ownerID, leadID int64) ([]domain.Touchpoint, error) {
	return s.repo.ListTouchpoints(ctx, ownerID, leadID)
}

func (s *Service) CreateTouchpoint(ctx context.Context, ownerID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error) {
	if req.Type == "" {
		req.Type = "call"
	}
	if req.Outcome == "" {
		req.Outcome = "other"
	}
	return s.repo.CreateTouchpoint(ctx, ownerID, createdBy, req)
}

func (s *Service) DeleteTouchpoint(ctx context.Context, id, ownerID int64) error {
	return s.repo.DeleteTouchpoint(ctx, id, ownerID)
}

func (s *Service) ListContracts(ctx context.Context, ownerID int64) ([]domain.Contract, error) {
	return s.repo.ListContracts(ctx, ownerID)
}

func (s *Service) GetContract(ctx context.Context, id, ownerID int64) (domain.Contract, error) {
	return s.repo.GetContract(ctx, id, ownerID)
}

func (s *Service) CreateContract(ctx context.Context, ownerID int64, req domain.CreateContractRequest) (domain.Contract, error) {
	if req.Plan == "" {
		req.Plan = "pro"
	}
	if req.Seats == 0 {
		req.Seats = 1
	}
	return s.repo.CreateContract(ctx, ownerID, req)
}

func (s *Service) UpdateContract(ctx context.Context, id, ownerID int64, req domain.UpdateContractRequest) (domain.Contract, error) {
	if req.Plan == "" {
		req.Plan = "pro"
	}
	if req.Seats == 0 {
		req.Seats = 1
	}
	return s.repo.UpdateContract(ctx, id, ownerID, req)
}
