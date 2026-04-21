package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64, dateFrom, dateTo string) ([]domain.Receipt, error) {
	var receipts []domain.Receipt

	if dateFrom != "" && dateTo != "" {
		rows, err := r.queries.ListReceiptsWithDateRange(ctx, sqlc.ListReceiptsWithDateRangeParams{
			OwnerID: ownerID,
			PaidAt:  sql.NullTime{Time: parseTime(dateFrom), Valid: true},
			PaidAt_2: sql.NullTime{Time: parseTime(dateTo), Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		for _, row := range rows {
			receipts = append(receipts, domain.Receipt{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				CustomerName:  row.CustomerName.String,
				Total:         decimalToFloat(row.Total.String),
				PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
				PaidAt:        row.PaidAt.Time,
			})
		}
	} else if dateFrom != "" {
		rows, err := r.queries.ListReceiptsWithDateFrom(ctx, sqlc.ListReceiptsWithDateFromParams{
			OwnerID: ownerID,
			PaidAt:  sql.NullTime{Time: parseTime(dateFrom), Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		for _, row := range rows {
			receipts = append(receipts, domain.Receipt{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				CustomerName:  row.CustomerName.String,
				Total:         decimalToFloat(row.Total.String),
				PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
				PaidAt:        row.PaidAt.Time,
			})
		}
	} else {
		rows, err := r.queries.ListReceipts(ctx, ownerID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		for _, row := range rows {
			receipts = append(receipts, domain.Receipt{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				CustomerName:  row.CustomerName.String,
				Total:         decimalToFloat(row.Total.String),
				PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
				PaidAt:        row.PaidAt.Time,
			})
		}
	}

	if receipts == nil {
		receipts = make([]domain.Receipt, 0)
	}
	return receipts, nil
}

func (r *Repository) Get(ctx context.Context, ownerID, id int64) (*domain.Receipt, error) {
	row, err := r.queries.GetReceiptByID(ctx, sqlc.GetReceiptByIDParams{ID: id, OwnerID: ownerID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReceiptNotFound
		}
		logger.Ctx(ctx).Error("Failed to get receipt", logger.Error(err))
		return nil, err
	}
	return &domain.Receipt{
		ID:            row.ID,
		OrderNumber:   row.OrderNumber,
		CustomerName:  row.CustomerName.String,
		Total:         decimalToFloat(row.Total.String),
		PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
		PaidAt:        row.PaidAt.Time,
	}, nil
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseTime(s string) (t time.Time) {
	t, _ = time.Parse("2006-01-02", s)
	return
}
