package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) List(ctx context.Context, businessID int64, dateFrom, dateTo string) ([]domain.Receipt, error) {
	var receipts []domain.Receipt

	if dateFrom != "" && dateTo != "" {
		end := parseTime(dateTo).AddDate(0, 0, 1)
		rows, err := r.queries.ListReceiptsWithDateRange(ctx, sqlc.ListReceiptsWithDateRangeParams{
			BusinessID: businessID,
			PaidAt:     sql.NullTime{Time: parseTime(dateFrom), Valid: true},
			PaidAt_2:   sql.NullTime{Time: end, Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		receipts = make([]domain.Receipt, 0, len(rows))
		for _, row := range rows {
			receipts = append(receipts, domain.Receipt{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				CustomerName:  row.CustomerName.String,
				Total:         sqlutil.DecimalToFloat(row.Total.String),
				PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
				PaidAt:        row.PaidAt.Time,
				BusinessID:    row.BusinessID,
			})
		}
	} else if dateFrom != "" {
		rows, err := r.queries.ListReceiptsWithDateFrom(ctx, sqlc.ListReceiptsWithDateFromParams{
			BusinessID: businessID,
			PaidAt:     sql.NullTime{Time: parseTime(dateFrom), Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		receipts = make([]domain.Receipt, 0, len(rows))
		for _, row := range rows {
			receipts = append(receipts, domain.Receipt{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				CustomerName:  row.CustomerName.String,
				Total:         sqlutil.DecimalToFloat(row.Total.String),
				PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
				PaidAt:        row.PaidAt.Time,
				BusinessID:    row.BusinessID,
			})
		}
	} else {
		rows, err := r.queries.ListReceipts(ctx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		receipts = make([]domain.Receipt, 0, len(rows))
		for _, row := range rows {
			receipts = append(receipts, domain.Receipt{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				CustomerName:  row.CustomerName.String,
				Total:         sqlutil.DecimalToFloat(row.Total.String),
				PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
				PaidAt:        row.PaidAt.Time,
				BusinessID:    row.BusinessID,
			})
		}
	}

	if receipts == nil {
		receipts = make([]domain.Receipt, 0)
	}
	return receipts, nil
}

func (r *Repository) Get(ctx context.Context, businessID, id int64) (*domain.Receipt, error) {
	row, err := r.queries.GetReceiptByID(ctx, sqlc.GetReceiptByIDParams{ID: id, BusinessID: businessID})
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
		Total:         sqlutil.DecimalToFloat(row.Total.String),
		PaymentMethod: string(row.PaymentMethod.PosOrdersPaymentMethod),
		PaidAt:        row.PaidAt.Time,
		BusinessID:    row.BusinessID,
	}, nil
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		logger.Global().Warn("failed to parse receipt date filter",
			logger.String("value", s), logger.Error(err))
		return time.Time{}
	}
	return t
}
