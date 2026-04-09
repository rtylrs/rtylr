package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) ListTransactions(ctx context.Context, ownerID int64, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	countQuery := "SELECT COUNT(*) FROM crm_loyalty_transactions WHERE owner_id = ?"
	args := []interface{}{ownerID}
	if customerID > 0 {
		countQuery += " AND customer_id = ?"
		args = append(args, customerID)
	}
	if err := r.pool.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := "SELECT id, owner_id, customer_id, points, type, description, reference_id, reference_type, created_at FROM crm_loyalty_transactions WHERE owner_id = ?"
	if customerID > 0 {
		query += " AND customer_id = ?"
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	queryArgs := append(args, pageSize, offset)
	rows, err := r.pool.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	txs := make([]domain.LoyaltyTransaction, 0)
	for rows.Next() {
		var tx domain.LoyaltyTransaction
		var desc sql.NullString
		var refID sql.NullInt64
		var refType sql.NullString
		if err := rows.Scan(&tx.ID, &tx.OwnerID, &tx.CustomerID, &tx.Points, &tx.Type, &desc, &refID, &refType, &tx.CreatedAt); err != nil {
			return nil, 0, err
		}
		tx.Description = desc.String
		if refID.Valid {
			tx.ReferenceID = &refID.Int64
		}
		tx.ReferenceType = refType.String
		txs = append(txs, tx)
	}
	return txs, total, rows.Err()
}

func (r *Repository) CreateTransaction(ctx context.Context, input domain.AwardInput, txType string) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_transactions (owner_id, customer_id, points, type, description, reference_id, reference_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.CustomerID, input.Points, txType,
		nullableString(input.Description), input.ReferenceID, nullableString(input.ReferenceType),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
	}
	return err
}

func (r *Repository) CreateRedeemTransaction(ctx context.Context, input domain.RedeemInput) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_transactions (owner_id, customer_id, points, type, description, reference_id, reference_type)
		 VALUES (?, ?, ?, 'redeem', ?, ?, ?)`,
		input.OwnerID, input.CustomerID, input.Points,
		nullableString(input.Description), input.ReferenceID, nullableString(input.ReferenceType),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create redeem transaction", logger.Error(err))
	}
	return err
}

func (r *Repository) GetConfig(ctx context.Context, ownerID int64) (domain.LoyaltyConfig, error) {
	var cfg domain.LoyaltyConfig
	err := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, points_per_dollar, min_redemption_points, points_value, is_active, created_at, updated_at
		 FROM crm_loyalty_config WHERE owner_id = ?`, ownerID,
	).Scan(&cfg.ID, &cfg.OwnerID, &cfg.PointsPerDollar, &cfg.MinRedemptionPoints, &cfg.PointsValue, &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, domain.ErrLoyaltyConfigNotFound
		}
		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))
		return domain.LoyaltyConfig{}, err
	}
	return cfg, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_config (owner_id, points_per_dollar, min_redemption_points, points_value, is_active)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   points_per_dollar = VALUES(points_per_dollar),
		   min_redemption_points = VALUES(min_redemption_points),
		   points_value = VALUES(points_value),
		   is_active = VALUES(is_active)`,
		input.OwnerID, input.PointsPerDollar, input.MinRedemptionPoints, input.PointsValue, input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}
	return err
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
