package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) CreateDocument(ctx context.Context, businessID int64, params domain.CreateDocumentParams) (*domain.EmployeeDocumentRecord, error) {
	if err := r.ensureEmployeeBelongsToBusiness(ctx, businessID, params.EmployeeID); err != nil {
		return nil, err
	}

	result, err := r.queries.CreateDocument(ctx, sqlc.CreateDocumentParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		EmployeeID:   params.EmployeeID,
		DocumentType: params.DocumentType,
		Title:        params.Title,
		StorageMode:  params.StorageMode,
		FileUrl:      sqlutil.NullString(params.FileURL),
		ExternalUrl:  sqlutil.NullString(params.ExternalURL),
		ExpiryDate:   nullableTime(params.ExpiryDate),
		UploadedBy:   params.UploadedBy,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetDocument(ctx, businessID, id)
}

func (r *Repository) GetDocument(ctx context.Context, businessID, id int64) (*domain.EmployeeDocumentRecord, error) {
	row, err := r.queries.GetDocument(ctx, sqlc.GetDocumentParams{
		BusinessID: businessID,
		ID:         id,
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

func (r *Repository) ListDocuments(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	limit, offset = normalizeDocumentPage(limit, offset)
	total, err := r.queries.CountDocuments(ctx, sqlc.CountDocumentsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListDocuments(ctx, sqlc.ListDocumentsParams{
		BusinessID: businessID,
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
	if err := r.populateEmployeeNames(ctx, businessID, records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *Repository) DeleteDocument(ctx context.Context, businessID, id int64) error {
	result, err := r.queries.DeleteDocument(ctx, sqlc.DeleteDocumentParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrDocumentNotFound
	}
	return nil
}

// ensureEmployeeBelongsToBusiness prevents the document table's intentionally
// loose employee_id column from becoming a cross-tenant association.  The
// check is deliberately scoped by both identifiers; an id alone is not a
// tenant boundary.
func (r *Repository) ensureEmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) error {
	if businessID <= 0 || employeeID <= 0 {
		return domain.ErrDocumentNotFound
	}
	if r.db == nil {
		return errors.New("document repository database is unavailable")
	}

	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM hr_employees
		WHERE id = ? AND business_id = ?
		LIMIT 1`, employeeID, businessID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("employee %d is not in business %d: %w", employeeID, businessID, domain.ErrDocumentNotFound)
	}
	return err
}

func nullableTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func (r *Repository) populateEmployeeNames(ctx context.Context, businessID int64, records []domain.EmployeeDocumentRecord) error {
	if len(records) == 0 {
		return nil
	}
	if r.db == nil {
		return errors.New("document repository database is unavailable")
	}

	ids := make([]int64, 0, len(records))
	seen := make(map[int64]struct{}, len(records))
	for _, record := range records {
		if _, ok := seen[record.EmployeeID]; ok {
			continue
		}
		seen[record.EmployeeID] = struct{}{}
		ids = append(ids, record.EmployeeID)
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+1)
	args = append(args, businessID)
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, TRIM(CONCAT(first_name, ' ', last_name))
		FROM hr_employees
		WHERE business_id = ? AND id IN (`+placeholders+`)`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	names := make(map[int64]string, len(ids))
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		names[id] = name
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range records {
		records[i].EmployeeName = names[records[i].EmployeeID]
	}
	return nil
}

func normalizeDocumentPage(limit, offset int32) (int32, int32) {
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

func documentFromRow(row sqlc.HrEmployeeDocument) domain.EmployeeDocumentRecord {
	var expiryDate *time.Time
	if row.ExpiryDate.Valid {
		t := row.ExpiryDate.Time
		expiryDate = &t
	}
	return domain.EmployeeDocumentRecord{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		BusinessID:   row.BusinessID,
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

func (r *Repository) ListExpiringDocuments(ctx context.Context, businessID int64, days, limit, offset int32) ([]domain.EmployeeDocumentRecord, int64, error) {
	limit, offset = normalizeDocumentPage(limit, offset)
	cutoff := time.Now().AddDate(0, 0, int(days))

	total, err := r.queries.CountExpiringDocuments(ctx, sqlc.CountExpiringDocumentsParams{
		BusinessID: businessID,
		ExpiryDate: sql.NullTime{Time: cutoff, Valid: true},
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListExpiringDocuments(ctx, sqlc.ListExpiringDocumentsParams{
		BusinessID: businessID,
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
	if err := r.populateEmployeeNames(ctx, businessID, records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
