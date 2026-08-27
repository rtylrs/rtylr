package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrDocumentsScope = "hr_documents"
	hrDocumentsTTL   = 5 * time.Minute
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateHRDocumentsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrDocumentsScope, businessID))
}

func (s *Service) UploadDocument(ctx context.Context, businessID int64, params domain.CreateDocumentParams) (*domain.EmployeeDocumentRecord, error) {
	if params.StorageMode != "stored" && params.StorageMode != "linked" {
		return nil, domain.ErrStorageModeInvalid
	}
	if params.StorageMode == "stored" && params.FileURL == "" {
		return nil, domain.ErrStorageModeInvalid
	}
	if params.StorageMode == "linked" && params.ExternalURL == "" {
		return nil, domain.ErrStorageModeInvalid
	}

	rec, err := s.repo.CreateDocument(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHRDocumentsScope(ctx, businessID)
	return rec, nil
}

type listDocumentsResult struct {
	Records []domain.EmployeeDocumentRecord
	Total   int64
}

func (s *Service) ListDocuments(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	key := cacheKeys.Scoped(hrDocumentsScope, businessID, fmt.Sprintf("list|e=%d|l=%d|o=%d", employeeID, limit, offset))
	res, err := cacher.Of[listDocumentsResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (listDocumentsResult, error) {
			recs, total, err := s.repo.ListDocuments(ctx, businessID, employeeID, limit, offset)
			return listDocumentsResult{Records: recs, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrDocumentsTTL),
	)
	return res.Records, res.Total, err
}

func (s *Service) ListExpiringDocuments(ctx context.Context, businessID int64, days, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	key := cacheKeys.Scoped(hrDocumentsScope, businessID, fmt.Sprintf("exp|d=%d|l=%d|o=%d", days, limit, offset))
	res, err := cacher.Of[listDocumentsResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (listDocumentsResult, error) {
			recs, total, err := s.repo.ListExpiringDocuments(ctx, businessID, days, limit, offset)
			return listDocumentsResult{Records: recs, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrDocumentsTTL),
	)
	return res.Records, res.Total, err
}

func (s *Service) GetDocument(ctx context.Context, businessID, id int64) (*domain.EmployeeDocumentRecord, error) {
	key := cacheKeys.Scoped(hrDocumentsScope, businessID, fmt.Sprintf("id=%d", id))
	return cacher.Of[*domain.EmployeeDocumentRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.EmployeeDocumentRecord, error) {
			return s.repo.GetDocument(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, hrDocumentsTTL),
	)
}

func (s *Service) DeleteDocument(ctx context.Context, businessID, id int64) error {
	if _, err := s.repo.GetDocument(ctx, businessID, id); err != nil {
		return err
	}
	if err := s.repo.DeleteDocument(ctx, businessID, id); err != nil {
		return err
	}
	invalidateHRDocumentsScope(ctx, businessID)
	return nil
}
