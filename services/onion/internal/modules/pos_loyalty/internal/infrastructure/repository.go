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
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
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
