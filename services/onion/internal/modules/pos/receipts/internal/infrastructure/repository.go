package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/pos/receipts/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/receipts/internal/infrastructure/persistence/sqlc"
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

func mapReceiptRow(
	id int64,
	orderNumber string,
	customerName sql.NullString,
	total sql.NullString,
	paymentMethod sqlc.NullPosOrdersPaymentMethod,
	paidAt sql.NullTime,
	createdAt sql.NullTime,
	businessID int64,
) domain.Receipt {
	timestamp := paidAt.Time
	if !paidAt.Valid {
		timestamp = createdAt.Time
	}
	method := ""
	if paymentMethod.Valid {
		method = string(paymentMethod.PosOrdersPaymentMethod)
	}
	if method == "" {
		method = "online"
	}
	return domain.Receipt{
		ID:            id,
		OrderNumber:   orderNumber,
		CustomerName:  customerName.String,
		Total:         sqlutil.DecimalToFloat(total.String),
		PaymentMethod: method,
		PaidAt:        timestamp,
		BusinessID:    businessID,
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
			receipts = append(receipts, mapReceiptRow(
				row.ID, row.OrderNumber, row.CustomerName, row.Total,
				row.PaymentMethod, row.PaidAt, row.CreatedAt, row.BusinessID,
			))
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
			receipts = append(receipts, mapReceiptRow(
				row.ID, row.OrderNumber, row.CustomerName, row.Total,
				row.PaymentMethod, row.PaidAt, row.CreatedAt, row.BusinessID,
			))
		}
	} else if dateTo != "" {
		end := parseTime(dateTo).AddDate(0, 0, 1)
		rows, err := r.queries.ListReceiptsWithDateTo(ctx, sqlc.ListReceiptsWithDateToParams{
			BusinessID: businessID,
			PaidAt:     sql.NullTime{Time: end, Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		receipts = make([]domain.Receipt, 0, len(rows))
		for _, row := range rows {
			receipts = append(receipts, mapReceiptRow(
				row.ID, row.OrderNumber, row.CustomerName, row.Total,
				row.PaymentMethod, row.PaidAt, row.CreatedAt, row.BusinessID,
			))
		}
	} else {
		rows, err := r.queries.ListReceipts(ctx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
			return nil, err
		}
		receipts = make([]domain.Receipt, 0, len(rows))
		for _, row := range rows {
			receipts = append(receipts, mapReceiptRow(
				row.ID, row.OrderNumber, row.CustomerName, row.Total,
				row.PaymentMethod, row.PaidAt, row.CreatedAt, row.BusinessID,
			))
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
	receipt := mapReceiptRow(
		row.ID, row.OrderNumber, row.CustomerName, row.Total,
		row.PaymentMethod, row.PaidAt, row.CreatedAt, row.BusinessID,
	)
	return &receipt, nil
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
