package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64, dateFrom, dateTo string) ([]domain.Receipt, error) {
	query := `SELECT id, order_number, COALESCE(customer_name,''), total, payment_method, paid_at
		 FROM pos_orders
		 WHERE owner_id = ? AND payment_status = 'paid'`

	args := []interface{}{ownerID}

	if dateFrom != "" {
		query += ` AND paid_at >= ?`
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		query += ` AND paid_at <= ?`
		args = append(args, dateTo)
	}

	query += ` ORDER BY paid_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list receipts", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	receipts := make([]domain.Receipt, 0)
	for rows.Next() {
		var rc domain.Receipt
		if err := rows.Scan(&rc.ID, &rc.OrderNumber, &rc.CustomerName, &rc.Total, &rc.PaymentMethod, &rc.PaidAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan receipt", logger.Error(err))
			return nil, err
		}
		receipts = append(receipts, rc)
	}
	return receipts, rows.Err()
}
