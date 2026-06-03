package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		db:      pool,
		queries: sqlc.New(pool),
	}
}

func (r *Repository) CreateDocument(ctx context.Context, ownerID int64, params domain.CreateDocumentParams) (*domain.EmployeeDocumentRecord, error) {
	result, err := r.queries.CreateDocument(ctx, sqlc.CreateDocumentParams{
		OwnerID:      ownerID,
		EmployeeID:   params.EmployeeID,
		DocumentType: params.DocumentType,
		Title:        params.Title,
		StorageMode:  params.StorageMode,
		FileUrl:      nullString(params.FileURL),
		ExternalUrl:  nullString(params.ExternalURL),
		UploadedBy:   params.UploadedBy,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetDocument(ctx, ownerID, id)
}

func (r *Repository) GetDocument(ctx context.Context, ownerID, id int64) (*domain.EmployeeDocumentRecord, error) {
	row, err := r.queries.GetDocument(ctx, sqlc.GetDocumentParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDocumentNotFound
		}
		return nil, err
	}

	record := documentFromRow(row)
	return &record, nil
}

func (r *Repository) ListDocuments(ctx context.Context, ownerID, employeeID int64, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	total, err := r.queries.CountDocuments(ctx, sqlc.CountDocumentsParams{
		OwnerID:    ownerID,
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListDocuments(ctx, sqlc.ListDocumentsParams{
		OwnerID:    ownerID,
		EmployeeID: employeeID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.EmployeeDocumentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, documentFromRow(row))
	}

	return records, total, nil
}

func (r *Repository) DeleteDocument(ctx context.Context, ownerID, id int64) error {
	return r.queries.DeleteDocument(ctx, sqlc.DeleteDocumentParams{
		OwnerID: ownerID,
		ID:      id,
	})
}

func documentFromRow(row sqlc.HrEmployeeDocument) domain.EmployeeDocumentRecord {
	var expiryDate *time.Time
	if row.ExpiryDate.Valid {
		t := row.ExpiryDate.Time
		expiryDate = &t
	}
	return domain.EmployeeDocumentRecord{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		EmployeeID:   row.EmployeeID,
		DocumentType: row.DocumentType,
		Title:        row.Title,
		StorageMode:  row.StorageMode,
		FileURL:      row.FileUrl.String,
		ExternalURL:  row.ExternalUrl.String,
		ExpiryDate:   expiryDate,
		UploadedBy:   row.UploadedBy,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (r *Repository) ListExpiringDocuments(ctx context.Context, ownerID int64, days, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	cutoff := time.Now().AddDate(0, 0, int(days))

	total, err := r.queries.CountExpiringDocuments(ctx, sqlc.CountExpiringDocumentsParams{
		OwnerID:    ownerID,
		ExpiryDate: sql.NullTime{Time: cutoff, Valid: true},
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListExpiringDocuments(ctx, sqlc.ListExpiringDocumentsParams{
		OwnerID:    ownerID,
		ExpiryDate: sql.NullTime{Time: cutoff, Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.EmployeeDocumentRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, documentFromRow(row))
	}

	return records, total, nil
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}
