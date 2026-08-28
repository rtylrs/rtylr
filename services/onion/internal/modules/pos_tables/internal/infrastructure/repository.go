package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_tables/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Table, error) {
	rows, err := r.queries.ListTablesByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tables", logger.Error(err))
		return nil, err
	}

	tables := make([]domain.Table, 0, len(rows))
	for _, row := range rows {
		t := domain.Table{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Label:      row.Name,
			Seats:      int(row.Capacity.Int32),
			Status:     string(row.Status.PosTablesStatus),
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		}
		if row.LocationID.Valid {
			t.LocationID = &row.LocationID.Int64
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (r *Repository) Create(ctx context.Context, input domain.TableInput) (int64, error) {
	result, err := r.queries.CreateTable(ctx, sqlc.CreateTableParams{
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		LocationID: nullInt64Ptr(input.LocationID),
		Name:       input.Label,
		Capacity:   sql.NullInt32{Int32: int32(input.Seats), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create table", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TableUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	err := r.queries.UpdateTable(ctx, sqlc.UpdateTableParams{
		LocationID: nullInt64Ptr(input.LocationID),
		Name:       input.Label,
		Capacity:   sql.NullInt32{Int32: int32(input.Seats), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Status:     sqlc.NullPosTablesStatus{PosTablesStatus: sqlc.PosTablesStatus(input.Status), Valid: input.Status != ""},
		ID:         input.ID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update table", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, input domain.TableStatusInput) error {
	result, err := r.queries.UpdateTableStatus(ctx, sqlc.UpdateTableStatusParams{
		Status:     sqlc.NullPosTablesStatus{PosTablesStatus: sqlc.PosTablesStatus(input.Status), Valid: true},
		ID:         input.ID,
		BusinessID: input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update table status", logger.Error(err))
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTableNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	_, err := r.queries.DeleteTable(ctx, sqlc.DeleteTableParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete table", logger.Error(err))
		return err
	}
	return nil
}

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}
