package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.GiftCard, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, code, initial_amount, current_balance,
		        COALESCE(customer_name,''), COALESCE(customer_email,''),
		        is_active, expires_at, created_at, updated_at
		 FROM pos_gift_cards WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list gift cards", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.GiftCard, 0)
	for rows.Next() {
		gc, scanErr := scanGiftCard(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, gc)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.GiftCardInput) (int64, error) {
	var expiresAt sql.NullTime
	if input.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", input.ExpiresAt)
		if err == nil {
			expiresAt = sql.NullTime{Time: t, Valid: true}
		}
	}

	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO pos_gift_cards (owner_id, code, initial_amount, current_balance, customer_name, customer_email, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ownerID, input.Code, input.InitialAmount, input.InitialAmount,
		nullString(input.CustomerName), nullString(input.CustomerEmail), expiresAt,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create gift card", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetByCode(ctx context.Context, ownerID int64, code string) (*domain.GiftCard, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, code, initial_amount, current_balance,
		        COALESCE(customer_name,''), COALESCE(customer_email,''),
		        is_active, expires_at, created_at, updated_at
		 FROM pos_gift_cards WHERE owner_id = ? AND code = ?`, ownerID, code,
	)

	var gc domain.GiftCard
	var expiresAt sql.NullTime

	if err := row.Scan(
		&gc.ID, &gc.OwnerID, &gc.Code, &gc.InitialAmount, &gc.CurrentBalance,
		&gc.CustomerName, &gc.CustomerEmail, &gc.IsActive, &expiresAt,
		&gc.CreatedAt, &gc.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrGiftCardNotFound
		}
		logger.Ctx(ctx).Error("Failed to get gift card by code", logger.Error(err))
		return nil, err
	}

	if expiresAt.Valid {
		gc.ExpiresAt = &expiresAt.Time
	}
	return &gc, nil
}

func (r *Repository) Redeem(ctx context.Context, ownerID int64, input domain.RedeemInput) (*domain.GiftCard, error) {
	gc, err := r.GetByCode(ctx, ownerID, input.Code)
	if err != nil {
		return nil, err
	}
	if !gc.IsActive {
		return nil, domain.ErrGiftCardInactive
	}
	if gc.ExpiresAt != nil && gc.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrGiftCardExpired
	}
	if gc.CurrentBalance < input.Amount {
		return nil, domain.ErrInsufficientBalance
	}

	newBalance := gc.CurrentBalance - input.Amount
	_, err = r.pool.ExecContext(ctx,
		`UPDATE pos_gift_cards SET current_balance = ? WHERE id = ? AND owner_id = ?`,
		newBalance, gc.ID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to redeem gift card", logger.Error(err))
		return nil, err
	}

	gc.CurrentBalance = newBalance
	return gc, nil
}

func scanGiftCard(rows *sql.Rows) (domain.GiftCard, error) {
	var gc domain.GiftCard
	var expiresAt sql.NullTime
	if err := rows.Scan(
		&gc.ID, &gc.OwnerID, &gc.Code, &gc.InitialAmount, &gc.CurrentBalance,
		&gc.CustomerName, &gc.CustomerEmail, &gc.IsActive, &expiresAt,
		&gc.CreatedAt, &gc.UpdatedAt,
	); err != nil {
		return domain.GiftCard{}, err
	}
	if expiresAt.Valid {
		gc.ExpiresAt = &expiresAt.Time
	}
	return gc, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
