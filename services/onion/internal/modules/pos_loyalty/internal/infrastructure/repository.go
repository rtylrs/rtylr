package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	row, err := r.queries.GetLoyaltyConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, domain.ErrConfigNotFound
		}
		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))
		return domain.LoyaltyConfig{}, err
	}
	return domain.LoyaltyConfig{
		ID:              row.ID,
		BusinessID:      row.BusinessID,
		IsEnabled:       row.IsEnabled.Bool,
		PointsPerDollar: decimalToFloat(row.PointsPerDollar.String),
		RedemptionRate:  decimalToFloat(row.RedemptionRate.String),
		MinRedeemPoints: row.MinRedeemPoints.Int32,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.UpdateLoyaltyConfigInput) error {
	_, err := r.queries.UpsertLoyaltyConfig(ctx, sqlc.UpsertLoyaltyConfigParams{
		BusinessID:      input.BusinessID,
		IsEnabled:       sql.NullBool{Bool: input.IsEnabled, Valid: true},
		PointsPerDollar: sql.NullString{String: floatToDecimal(input.PointsPerDollar), Valid: true},
		RedemptionRate:  sql.NullString{String: fmt.Sprintf("%.4f", input.RedemptionRate), Valid: true},
		MinRedeemPoints: sql.NullInt32{Int32: input.MinRedeemPoints, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ListTransactions(ctx context.Context, businessID, customerID int64) ([]domain.LoyaltyTransaction, error) {
	rows, err := r.queries.ListLoyaltyTransactions(ctx, sqlc.ListLoyaltyTransactionsParams{
		BusinessID: businessID,
		CustomerID: customerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))
		return nil, err
	}
	txns := make([]domain.LoyaltyTransaction, 0, len(rows))
	for _, row := range rows {
		t := domain.LoyaltyTransaction{
			ID:           row.ID,
			BusinessID:   row.BusinessID,
			CustomerID:   row.CustomerID,
			Type:         string(row.Type),
			Points:       row.Points,
			BalanceAfter: row.BalanceAfter,
			CreatedAt:    row.CreatedAt.Time,
		}
		if row.OrderID.Valid {
			t.OrderID = &row.OrderID.Int64
		}
		if row.Note.Valid {
			t.Note = &row.Note.String
		}
		txns = append(txns, t)
	}
	return txns, nil
}

func (r *Repository) GetBalance(ctx context.Context, businessID, customerID int64) (int32, error) {
	raw, err := r.queries.GetLoyaltyBalance(ctx, sqlc.GetLoyaltyBalanceParams{
		BusinessID: businessID,
		CustomerID: customerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty balance", logger.Error(err))
		return 0, err
	}
	return int32(interfaceToInt(raw)), nil
}

// GetPaidOrderTotal is deliberately scoped by business, customer, and paid
// status. Loyalty awards must be derived from payment data, not client input.
func (r *Repository) GetPaidOrderTotal(ctx context.Context, businessID, customerID, orderID int64) (float64, error) {
	var total string
	err := r.pool.QueryRowContext(ctx, `
		SELECT total FROM pos_orders
		WHERE business_id = ? AND id = ? AND customer_id = ? AND payment_status = 'paid'
		LIMIT 1`, businessID, orderID, customerID).Scan(&total)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrOrderNotPaid
	}
	if err != nil {
		return 0, err
	}
	return decimalToFloat(total), nil
}

// CreateBalanceTransaction serializes the balance read/modify/write with a
// MySQL advisory lock. This also covers the first transaction for a customer,
// where a row-level FOR UPDATE would have nothing to lock.
func (r *Repository) CreateBalanceTransaction(ctx context.Context, businessID, customerID int64, orderID *int64, txnType string, points int32, note string) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	lockName := fmt.Sprintf("pos_loyalty:%d:%d", businessID, customerID)
	var acquired int
	if err := tx.QueryRowContext(ctx, `SELECT GET_LOCK(?, 5)`, lockName).Scan(&acquired); err != nil {
		return 0, err
	}
	if acquired != 1 {
		return 0, fmt.Errorf("could not lock loyalty balance")
	}
	defer tx.ExecContext(ctx, `SELECT RELEASE_LOCK(?)`, lockName)

	var current int32
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE((SELECT balance_after FROM pos_loyalty_transactions
		 WHERE business_id = ? AND customer_id = ? ORDER BY id DESC LIMIT 1), 0)`, businessID, customerID).Scan(&current); err != nil {
		return 0, err
	}
	if txnType == "earn" {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
			SELECT 1 FROM pos_loyalty_transactions
			WHERE business_id = ? AND customer_id = ? AND order_id = ? AND type = 'earn')`, businessID, customerID, *orderID).Scan(&exists); err != nil {
			return 0, err
		}
		if exists == 1 {
			return 0, domain.ErrOrderAlreadyEarned
		}
	} else if current+points < 0 {
		return 0, domain.ErrInsufficientPoints
	}
	newBalance := current + points
	var orderValue any
	if orderID != nil {
		orderValue = *orderID
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO pos_loyalty_transactions (business_id, customer_id, order_id, type, points, balance_after, note)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, businessID, customerID, orderValue, txnType, points, newBalance, note)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, businessID, customerID int64, orderID *int64, txnType string, points, balanceAfter int32, note *string) (int64, error) {
	result, err := r.queries.CreateLoyaltyTransaction(ctx, sqlc.CreateLoyaltyTransactionParams{
		BusinessID:   businessID,
		CustomerID:   customerID,
		OrderID:      nullInt64Ptr(orderID),
		Type:         sqlc.PosLoyaltyTransactionsType(txnType),
		Points:       points,
		BalanceAfter: balanceAfter,
		Note:         nullStringPtr(note),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func nullStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse loyalty decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

var floatToDecimal = sqlutil.FloatToDecimal

func interfaceToInt(v any) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	}
	var str string
	switch val := v.(type) {
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		str = fmt.Sprintf("%v", v)
	}
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logger.Global().Warn("failed to parse loyalty int from database",
			logger.String("value", str), logger.Error(err))
		return 0
	}
	return n
}
