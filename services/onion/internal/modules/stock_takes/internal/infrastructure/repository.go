package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListStockTakes(ctx context.Context, ownerID int64, limit, offset int) ([]domain.StockTake, bool, error) {
	const q = `
		SELECT id, owner_id, warehouse_id, status, notes, started_at, completed_at, created_at
		FROM erp_stock_takes
		WHERE owner_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.pool.QueryContext(ctx, q, ownerID, limit+1, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock takes", logger.Error(err))
		return nil, false, err
	}
	defer rows.Close()

	var stockTakes []domain.StockTake
	for rows.Next() {
		var s domain.StockTake
		var warehouseID sql.NullInt64
		var status, notes sql.NullString
		var startedAt, completedAt, createdAt sql.NullTime

		if err := rows.Scan(
			&s.ID, &s.OwnerID, &warehouseID, &status, &notes,
			&startedAt, &completedAt, &createdAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock take row", logger.Error(err))
			return nil, false, err
		}

		if warehouseID.Valid {
			s.WarehouseID = warehouseID.Int64
		}
		if status.Valid {
			s.Status = domain.StockTakeStatus(status.String)
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if startedAt.Valid {
			s.StartedAt = startedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if completedAt.Valid {
			s.CompletedAt = completedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if createdAt.Valid {
			s.CreatedAt = createdAt.Time
		}

		stockTakes = append(stockTakes, s)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(stockTakes) > limit
	if hasMore {
		stockTakes = stockTakes[:limit]
	}

	for i := range stockTakes {
		items, err := r.getStockTakeItems(ctx, stockTakes[i].ID)
		if err != nil {
			return nil, false, err
		}
		stockTakes[i].Items = items
	}

	return stockTakes, hasMore, nil
}

func (r *Repository) getStockTakeItems(ctx context.Context, stockTakeID int64) ([]domain.StockTakeItem, error) {
	rows, err := r.queries.GetStockTakeItems(ctx, stockTakeID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get stock take items", logger.Error(err))

		return nil, err
	}

	var items []domain.StockTakeItem

	for _, row := range rows {
		var item domain.StockTakeItem
		item.ID = row.ID
		item.StockTakeID = row.StockTakeID
		item.ProductID = row.ProductID
		item.ExpectedQty, _ = strconv.ParseFloat(row.ExpectedQty, 64)
		item.CountedQty, _ = strconv.ParseFloat(row.CountedQty, 64)
		item.Variance, _ = strconv.ParseFloat(row.Variance, 64)

		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) CreateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", logger.Error(err))

		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	res, err := qtx.InsertStockTake(ctx, sqlc.InsertStockTakeParams{
		OwnerID:     ownerID,
		WarehouseID: nullInt64(input.WarehouseID),
		Notes:       nullString(input.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock take", logger.Error(err))

		return 0, err
	}

	stockTakeID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range input.Items {
		variance := item.CountedQty - item.ExpectedQty
		_, err = qtx.InsertStockTakeItem(ctx, sqlc.InsertStockTakeItemParams{
			StockTakeID: stockTakeID,
			ProductID:   item.ProductID,
			ExpectedQty: strconv.FormatFloat(item.ExpectedQty, 'f', -1, 64),
			CountedQty:  strconv.FormatFloat(item.CountedQty, 'f', -1, 64),
			Variance:    strconv.FormatFloat(variance, 'f', -1, 64),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create stock take item", logger.Error(err))

			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take", logger.Error(err))

		return 0, err
	}

	return stockTakeID, nil
}

func (r *Repository) UpdateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeUpdateInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", logger.Error(err))

		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	statusVal := sqlc.NullErpStockTakesStatus{
		ErpStockTakesStatus: sqlc.ErpStockTakesStatus(input.Status),
		Valid:               true,
	}

	switch input.Status {
	case domain.StockTakeStatusInProgress:
		err = qtx.UpdateStockTakeStatusInProgress(ctx, sqlc.UpdateStockTakeStatusInProgressParams{
			Status:  statusVal,
			Notes:   nullString(input.Notes),
			OwnerID: ownerID,
			ID:      input.ID,
		})
	case domain.StockTakeStatusCompleted:
		err = qtx.UpdateStockTakeStatusCompleted(ctx, sqlc.UpdateStockTakeStatusCompletedParams{
			Status:  statusVal,
			Notes:   nullString(input.Notes),
			OwnerID: ownerID,
			ID:      input.ID,
		})
	default:
		err = qtx.UpdateStockTakeStatus(ctx, sqlc.UpdateStockTakeStatusParams{
			Status:  statusVal,
			Notes:   nullString(input.Notes),
			OwnerID: ownerID,
			ID:      input.ID,
		})
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock take", logger.Error(err))

		return err
	}

	if len(input.Items) > 0 {
		err = qtx.DeleteStockTakeItems(ctx, input.ID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to clear stock take items", logger.Error(err))

			return err
		}

		for _, item := range input.Items {
			variance := item.CountedQty - item.ExpectedQty
			_, err = qtx.InsertStockTakeItem(ctx, sqlc.InsertStockTakeItemParams{
				StockTakeID: input.ID,
				ProductID:   item.ProductID,
				ExpectedQty: strconv.FormatFloat(item.ExpectedQty, 'f', -1, 64),
				CountedQty:  strconv.FormatFloat(item.CountedQty, 'f', -1, 64),
				Variance:    strconv.FormatFloat(variance, 'f', -1, 64),
			})
			if err != nil {
				logger.Ctx(ctx).Error("Failed to insert stock take item", logger.Error(err))

				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take update", logger.Error(err))

		return err
	}

	return nil
}

func stockTakeFromRow(row sqlc.ErpStockTake) domain.StockTake {
	var s domain.StockTake
	s.ID = row.ID
	s.OwnerID = row.OwnerID

	if row.WarehouseID.Valid {
		s.WarehouseID = row.WarehouseID.Int64
	}

	if row.Status.Valid {
		s.Status = domain.StockTakeStatus(row.Status.ErpStockTakesStatus)
	}

	if row.Notes.Valid {
		s.Notes = row.Notes.String
	}

	if row.StartedAt.Valid {
		s.StartedAt = row.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	if row.CompletedAt.Valid {
		s.CompletedAt = row.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	if row.CreatedAt.Valid {
		s.CreatedAt = row.CreatedAt.Time
	}

	return s
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
