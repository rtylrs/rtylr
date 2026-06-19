package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.GiftCard, error) {
	rows, err := r.queries.ListGiftCardsByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list gift cards", logger.Error(err))
		return nil, err
	}

	items := make([]domain.GiftCard, 0, len(rows))
	for _, row := range rows {
		gc := domain.GiftCard{
			ID:             row.ID,
			OwnerID:        row.OwnerID,
			BusinessID:     row.BusinessID,
			CustomerID:     nullableInt64(row.CustomerID),
			Code:           row.Code,
			InitialAmount:  decimalToFloat(row.InitialAmount),
			CurrentBalance: decimalToFloat(row.CurrentBalance),
			CustomerName:   row.CustomerName.String,
			CustomerEmail:  row.CustomerEmail.String,
			IsActive:       row.IsActive.Bool,
			CreatedAt:      row.CreatedAt.Time,
			UpdatedAt:      row.UpdatedAt.Time,
		}
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time
			gc.ExpiresAt = &t
		}
		items = append(items, gc)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.GiftCardInput) (int64, error) {
	var expiresAt sql.NullTime
	if input.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", input.ExpiresAt)
		if err == nil {
			expiresAt = sql.NullTime{Time: t, Valid: true}
		}
	}

	amountStr := floatToDecimal(input.InitialAmount)

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin gift card transaction", logger.Error(err))
		return 0, err
	}
	defer tx.Rollback()

	customerID, err := customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sync gift card customer", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateGiftCard(ctx, sqlc.CreateGiftCardParams{
		OwnerID:        0,
		BusinessID:     businessID,
		CustomerID:     nullableInt64Value(customerID),
		Code:           input.Code,
		InitialAmount:  amountStr,
		CurrentBalance: amountStr,
		CustomerName:   sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail:  sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create gift card", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit gift card transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) GetByCode(ctx context.Context, businessID int64, code string) (*domain.GiftCard, error) {
	row, err := r.queries.GetGiftCardByCode(ctx, sqlc.GetGiftCardByCodeParams{
		BusinessID: businessID,
		Code:       code,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGiftCardNotFound
		}
		logger.Ctx(ctx).Error("Failed to get gift card by code", logger.Error(err))
		return nil, err
	}

	return giftCardFromRow(row), nil
}

func (r *Repository) Redeem(ctx context.Context, businessID int64, input domain.RedeemInput) (*domain.GiftCard, error) {
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin redeem transaction",
			logger.Int("business_id", int(businessID)), logger.String("code", input.Code), logger.Error(err))
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	// Lock the row for the duration of the transaction so concurrent redemptions
	// serialise on this gift card and cannot double-spend the balance.
	row, err := qtx.GetGiftCardByCodeForUpdate(ctx, sqlc.GetGiftCardByCodeForUpdateParams{
		BusinessID: businessID,
		Code:       input.Code,
	})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGiftCardNotFound
		}
		logger.Ctx(ctx).Error("Failed to lock gift card for redeem",
			logger.Int("business_id", int(businessID)), logger.String("code", input.Code), logger.Error(err))
		return nil, err
	}

	gc := giftCardFromRow(row)

	if !gc.IsActive {
		_ = tx.Rollback()
		return nil, domain.ErrGiftCardInactive
	}
	if gc.ExpiresAt != nil && gc.ExpiresAt.Before(time.Now()) {
		_ = tx.Rollback()
		return nil, domain.ErrGiftCardExpired
	}
	if gc.CurrentBalance < input.Amount {
		_ = tx.Rollback()
		return nil, domain.ErrInsufficientBalance
	}

	newBalance := gc.CurrentBalance - input.Amount

	if err := qtx.UpdateGiftCardBalance(ctx, sqlc.UpdateGiftCardBalanceParams{
		CurrentBalance: floatToDecimal(newBalance),
		ID:             gc.ID,
		BusinessID:     businessID,
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to redeem gift card",
			logger.Int("business_id", int(businessID)), logger.String("code", input.Code), logger.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit redeem transaction",
			logger.Int("business_id", int(businessID)), logger.String("code", input.Code), logger.Error(err))
		return nil, err
	}

	gc.CurrentBalance = newBalance
	return gc, nil
}

// giftCardFromRow maps a sqlc PosGiftCard row to the domain GiftCard.
func giftCardFromRow(row sqlc.PosGiftCard) *domain.GiftCard {
	gc := &domain.GiftCard{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		CustomerID:     nullableInt64(row.CustomerID),
		Code:           row.Code,
		InitialAmount:  decimalToFloat(row.InitialAmount),
		CurrentBalance: decimalToFloat(row.CurrentBalance),
		CustomerName:   row.CustomerName.String,
		CustomerEmail:  row.CustomerEmail.String,
		IsActive:       row.IsActive.Bool,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		gc.ExpiresAt = &t
	}
	return gc
}

func nullableInt64(v sql.NullInt64) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}

func nullableInt64Value(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse gift card decimal value from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
