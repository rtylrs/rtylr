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
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.DocumentInput) (int64, error) {
	res, err := r.queries.CreateDocument(ctx, sqlc.CreateDocumentParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		StaffID:    sql.NullInt64{Int64: input.StaffID, Valid: input.StaffID != 0},
		Name:       input.Name,
		Type:       sql.NullString{String: input.Type, Valid: input.Type != ""},
		FileUrl:    sql.NullString{String: input.FileURL, Valid: input.FileURL != ""},
		ExpiryDate: parseNullTime(input.ExpiryDate),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create document", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteDocument(ctx, sqlc.DeleteDocumentParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete document", logger.Error(err))
		return err
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
