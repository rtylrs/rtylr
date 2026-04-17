package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) UploadDocument(ctx context.Context, ownerID int64, params domain.CreateDocumentParams) (*domain.EmployeeDocumentRecord, error) {
	if params.StorageMode != "stored" && params.StorageMode != "linked" {
		return nil, domain.ErrStorageModeInvalid
	}

	if params.StorageMode == "stored" && params.FileURL == "" {
		return nil, domain.ErrStorageModeInvalid
	}

	if params.StorageMode == "linked" && params.ExternalURL == "" {
		return nil, domain.ErrStorageModeInvalid
	}

	return s.repo.CreateDocument(ctx, ownerID, params)
}

func (s *Service) ListDocuments(ctx context.Context, ownerID, employeeID int64, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListDocuments(ctx, ownerID, employeeID, limit, offset)
}

func (s *Service) ListExpiringDocuments(ctx context.Context, ownerID int64, days, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListExpiringDocuments(ctx, ownerID, days, limit, offset)
}

func (s *Service) GetDocument(ctx context.Context, ownerID, id int64) (*domain.EmployeeDocumentRecord, error) {
	return s.repo.GetDocument(ctx, ownerID, id)
}

func (s *Service) DeleteDocument(ctx context.Context, ownerID, id int64) error {
	_, err := s.repo.GetDocument(ctx, ownerID, id)
	if err != nil {
		return err
	}

	return s.repo.DeleteDocument(ctx, ownerID, id)
}
