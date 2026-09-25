package application

import (
	"context"
	"database/sql"
	"errors"
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
	db   *sql.DB
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
		db:   db,
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

	// Keep the repository path for the db-less unit-test seam, but production
	// services use the transactional path below so a readback failure cannot
	// leave a committed row behind an error response.
	if s.db == nil {
		if s.repo == nil {
			return nil, errors.New("document repository database is unavailable")
		}
		rec, err := s.repo.CreateDocument(ctx, businessID, params)
		if err != nil {
			invalidateHRDocumentsScope(context.WithoutCancel(ctx), businessID)
			return nil, err
		}
		invalidateHRDocumentsScope(ctx, businessID)
		return rec, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var employeeExists int
	if err := tx.QueryRowContext(ctx, `
		SELECT 1
		FROM hr_employees
		WHERE id = ? AND business_id = ?
		LIMIT 1`, params.EmployeeID, businessID).Scan(&employeeExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, err
	}

	var fileURL any
	if params.FileURL != "" {
		fileURL = params.FileURL
	}
	var externalURL any
	if params.ExternalURL != "" {
		externalURL = params.ExternalURL
	}
	var insertExpiryDate any
	if params.ExpiryDate != nil {
		insertExpiryDate = *params.ExpiryDate
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO hr_employee_documents
		(owner_id, business_id, employee_id, document_type, title, storage_mode,
		 file_url, external_url, expiry_date, uploaded_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		businessID, businessID, params.EmployeeID, params.DocumentType, params.Title,
		params.StorageMode, fileURL, externalURL, insertExpiryDate, params.UploadedBy)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rec := &domain.EmployeeDocumentRecord{}
	var storedFileURL, storedExternalURL sql.NullString
	var expiryDate sql.NullTime
	if err := tx.QueryRowContext(ctx, `
		SELECT id, owner_id, business_id, employee_id, document_type, title,
		       storage_mode, file_url, external_url, expiry_date, uploaded_by,
		       created_at, updated_at
		FROM hr_employee_documents
		WHERE business_id = ? AND id = ?`, businessID, id).Scan(
		&rec.ID,
		&rec.OwnerID,
		&rec.BusinessID,
		&rec.EmployeeID,
		&rec.DocumentType,
		&rec.Title,
		&rec.StorageMode,
		&storedFileURL,
		&storedExternalURL,
		&expiryDate,
		&rec.UploadedBy,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return nil, err
	}
	rec.FileURL = storedFileURL.String
	rec.ExternalURL = storedExternalURL.String
	if expiryDate.Valid {
		t := expiryDate.Time
		rec.ExpiryDate = &t
	}

	if err := tx.Commit(); err != nil {
		// A commit error can be ambiguous at the network boundary. Even if
		// the server committed, do not leave a stale cached list behind.
		invalidateHRDocumentsScope(context.WithoutCancel(ctx), businessID)
		return nil, err
	}
	committed = true
	invalidateHRDocumentsScope(context.WithoutCancel(ctx), businessID)
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
	if offset < 0 {
		offset = 0
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
	if offset < 0 {
		offset = 0
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
	if s.repo == nil {
		return errors.New("document repository database is unavailable")
	}
	if err := s.repo.DeleteDocument(ctx, businessID, id); err != nil {
		return err
	}
	invalidateHRDocumentsScope(context.WithoutCancel(ctx), businessID)
	return nil
}
