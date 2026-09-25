package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	"github.com/voxire/onion/internal/modules/documents/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Document, error) {
	rows, err := r.queries.ListDocuments(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list documents", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Document, 0, len(rows))
	for _, row := range rows {
		items = append(items, documentFromRow(row))
	}
	return items, nil
}

// ListPaginated returns a single page of documents for the business, ordered
// by creation date. Limit and offset are applied as-is; callers clamp them.
func (r *Repository) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Document, error) {
	rows, err := r.queries.ListDocumentsPaginated(ctx, sqlc.ListDocumentsPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list documents page", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Document, 0, len(rows))
	for _, row := range rows {
		items = append(items, documentFromRow(row))
	}
	return items, nil
}

// documentFromRow maps a sqlc hr_documents row to the domain model,
// collapsing nullable columns to their zero value when NULL.
func documentFromRow(row sqlc.HrDocument) domain.Document {
	var item domain.Document
	item.ID = row.ID
	item.OwnerID = row.OwnerID
	item.BusinessID = row.BusinessID
	if row.StaffID.Valid {
		item.StaffID = row.StaffID.Int64
	}
	item.Name = row.Name
	if row.Type.Valid {
		item.Type = row.Type.String
	}
	if row.FileUrl.Valid {
		item.FileURL = row.FileUrl.String
	}
	if row.ExpiryDate.Valid {
		item.ExpiryDate = row.ExpiryDate.Time
	}
	if row.CreatedAt.Valid {
		item.CreatedAt = row.CreatedAt.Time
	}
	return item
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.DocumentInput) (int64, error) {
	var (
		res sql.Result
		err error
	)
	if input.StaffID == 0 {
		res, err = r.queries.CreateDocument(ctx, sqlc.CreateDocumentParams{
			OwnerID:    businessID,
			BusinessID: businessID,
			Name:       input.Name,
			Type:       sql.NullString{String: input.Type, Valid: input.Type != ""},
			FileUrl:    sql.NullString{String: input.FileURL, Valid: input.FileURL != ""},
			ExpiryDate: parseNullTime(input.ExpiryDate),
		})
	} else {
		res, err = r.queries.CreateDocumentForStaff(ctx, sqlc.CreateDocumentForStaffParams{
			OwnerID:         businessID,
			BusinessID:      businessID,
			StaffID:         input.StaffID,
			StaffBusinessID: businessID,
			Name:            input.Name,
			Type:            sql.NullString{String: input.Type, Valid: input.Type != ""},
			FileUrl:         sql.NullString{String: input.FileURL, Valid: input.FileURL != ""},
			ExpiryDate:      parseNullTime(input.ExpiryDate),
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create document", logger.Error(err))
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		if input.StaffID != 0 {
			return 0, domain.ErrDocumentStaffNotFound
		}
		return 0, sql.ErrNoRows
	}
	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	res, err := r.queries.DeleteDocument(ctx, sqlc.DeleteDocumentParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete document", logger.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrDocumentNotFound
	}
	return nil
}

func parseNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	for _, layout := range []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	return sql.NullTime{}
}
