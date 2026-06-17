package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	ownerID, err := r.ownerIDForBusiness(ctx, businessID)
	if err != nil {
		return domain.LoyaltyConfig{}, err
	}
	var row struct {
		id                  int64
		isEnabled           sql.NullBool
		pointsPerDollar     sql.NullString
		redemptionRate      sql.NullString
		minRedemptionPoints sql.NullInt32
		updatedAt           sql.NullTime
	}
	err = r.pool.QueryRowContext(ctx,
		`SELECT id, is_enabled, points_per_currency_unit, currency_per_point, min_redemption_points, updated_at
		 FROM crm_loyalty_config WHERE owner_id = ?`,
		ownerID).Scan(&row.id, &row.isEnabled, &row.pointsPerDollar, &row.redemptionRate, &row.minRedemptionPoints, &row.updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, domain.ErrConfigNotFound
		}
		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))
		return domain.LoyaltyConfig{}, err
	}
	return domain.LoyaltyConfig{
		ID:              row.id,
		BusinessID:      businessID,
		IsEnabled:       row.isEnabled.Bool,
		PointsPerDollar: decimalToFloat(row.pointsPerDollar.String),
		RedemptionRate:  decimalToFloat(row.redemptionRate.String),
		MinRedeemPoints: row.minRedemptionPoints.Int32,
		UpdatedAt:       row.updatedAt.Time,
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.UpdateLoyaltyConfigInput) error {
	ownerID, err := r.ownerIDForBusiness(ctx, input.BusinessID)
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_config
		 (owner_id, is_enabled, points_per_currency_unit, currency_per_point, min_redemption_points)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   is_enabled = VALUES(is_enabled),
		   points_per_currency_unit = VALUES(points_per_currency_unit),
		   currency_per_point = VALUES(currency_per_point),
		   min_redemption_points = VALUES(min_redemption_points)`,
		ownerID, input.IsEnabled, floatToDecimal(input.PointsPerDollar), fmt.Sprintf("%.4f", input.RedemptionRate), input.MinRedeemPoints)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ListTransactions(ctx context.Context, businessID, customerID int64) ([]domain.LoyaltyTransaction, error) {
	ownerID, err := r.ownerIDForBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, customer_id, reference_id, type, points, balance_after, description, created_at
		 FROM crm_loyalty_transactions
		 WHERE owner_id = ? AND customer_id = ?
		 ORDER BY created_at DESC`,
		ownerID, customerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	txns := make([]domain.LoyaltyTransaction, 0)
	for rows.Next() {
		var (
			id           int64
			cID          int64
			orderID      sql.NullInt64
			txnType      string
			points       int32
			balanceAfter int32
			note         sql.NullString
			createdAt    sql.NullTime
		)
		if err := rows.Scan(&id, &cID, &orderID, &txnType, &points, &balanceAfter, &note, &createdAt); err != nil {
			return nil, err
		}
		t := domain.LoyaltyTransaction{
			ID:           id,
			BusinessID:   businessID,
			CustomerID:   cID,
			Type:         txnType,
			Points:       points,
			BalanceAfter: balanceAfter,
			CreatedAt:    createdAt.Time,
		}
		if orderID.Valid {
			t.OrderID = &orderID.Int64
		}
		if note.Valid {
			t.Note = &note.String
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txns, nil
}

func (r *Repository) GetBalance(ctx context.Context, businessID, customerID int64) (int32, error) {
	ownerID, err := r.ownerIDForBusiness(ctx, businessID)
	if err != nil {
		return 0, err
	}
	var balance sql.NullInt32
	err = r.pool.QueryRowContext(ctx,
		`SELECT loyalty_points FROM crm_customers WHERE owner_id = ? AND id = ?`,
		ownerID, customerID).Scan(&balance)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get loyalty balance", logger.Error(err))
		return 0, err
	}
	if !balance.Valid {
		return 0, nil
	}
	return balance.Int32, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, businessID, customerID int64, orderID *int64, txnType string, points, balanceAfter int32, note *string) (int64, error) {
	ownerID, err := r.ownerIDForBusiness(ctx, businessID)
	if err != nil {
		return 0, err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	updateResult, err := tx.ExecContext(ctx,
		`UPDATE crm_customers
		 SET loyalty_points = ?, loyalty_tier = ?
		 WHERE owner_id = ? AND id = ?`,
		balanceAfter, tierForPoints(balanceAfter), ownerID, customerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update loyalty balance", logger.Error(err))
		return 0, err
	}
	updated, err := updateResult.RowsAffected()
	if err != nil {
		return 0, err
	}
	if updated == 0 {
		return 0, fmt.Errorf("loyalty customer %d not found for owner %d", customerID, ownerID)
	}

	var refType sql.NullString
	if orderID != nil {
		refType = sql.NullString{String: "pos_order", Valid: true}
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO crm_loyalty_transactions
		 (owner_id, customer_id, type, points, balance_before, balance_after, description, reference_id, reference_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ownerID,
		customerID,
		txnType,
		points,
		balanceAfter-points,
		balanceAfter,
		nullStringPtr(note),
		nullInt64Ptr(orderID),
		refType)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
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

func (r *Repository) ownerIDForBusiness(ctx context.Context, businessID int64) (int64, error) {
	var ownerID int64
	err := r.pool.QueryRowContext(ctx, `SELECT owner_id FROM businesses WHERE id = ?`, businessID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}

func tierForPoints(points int32) string {
	switch {
	case points >= 5000:
		return "platinum"
	case points >= 1500:
		return "gold"
	case points >= 500:
		return "silver"
	default:
		return "bronze"
	}
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

func floatToDecimal(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

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
