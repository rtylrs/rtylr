package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/voxire/onion/internal/modules/pos/internal/outbox"
	"math"
	"strings"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/giftcards/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
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

func (r *Repository) CreateForOwner(ctx context.Context, businessID, ownerID int64, input domain.GiftCardInput) (int64, error) {
	return r.create(ctx, businessID, ownerID, input)
}

func (r *Repository) create(ctx context.Context, businessID, ownerID int64, input domain.GiftCardInput) (int64, error) {
	var expiresAt sql.NullTime
	if input.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", input.ExpiresAt)
		if err == nil {
			expiresAt = sql.NullTime{Time: t, Valid: true}
		}
	}

	amountCents, err := dollarsToCents(input.InitialAmount)
	if err != nil {
		return 0, err
	}
	amountStr := amountCents.String()

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin gift card transaction", logger.Error(err))
		return 0, err
	}
	defer tx.Rollback()

	orderRef, err := lockPaidOrder(ctx, tx, businessID, input.PaidOrderID, amountCents)
	if err != nil {
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateGiftCard(ctx, sqlc.CreateGiftCardParams{
		OwnerID:        ownerID,
		BusinessID:     businessID,
		CustomerID:     sql.NullInt64{},
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
	contact := crmcontract.CustomerContactCaptured{Name: input.CustomerName, Email: input.CustomerEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "pos_gift_card",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("pos_gift_card", id, contact),
	}); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO pos_gift_card_transactions
			(gift_card_id, owner_id, business_id, type, order_ref, amount, balance_after, created_by)
		VALUES (?, ?, ?, 'issue', ?, ?, ?, ?)`,
		id, ownerID, businessID, orderRef, amountStr, amountStr, ownerID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit gift card issuance", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) LinkCustomerToGiftCard(ctx context.Context, businessID, giftCardID, customerID int64) error {
	return r.queries.LinkCustomerToGiftCard(ctx, sqlc.LinkCustomerToGiftCardParams{
		CustomerID: sql.NullInt64{Int64: customerID, Valid: true},
		BusinessID: businessID,
		ID:         giftCardID,
	})
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

func (r *Repository) RedeemForOrder(ctx context.Context, businessID, actorID int64, input domain.RedeemInput) (*domain.GiftCard, error) {
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin redeem transaction",
			logger.Int("business_id", int(businessID)), logger.String("code", domain.MaskCode(input.Code)), logger.Error(err))
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	amountCents, err := dollarsToCents(input.Amount)
	if err != nil {
		return nil, err
	}

	orderRef, err := lockGiftCardOrder(ctx, tx, businessID, input.OrderID, amountCents)
	if err != nil {
		return nil, err
	}

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
			logger.Int("business_id", int(businessID)), logger.String("code", domain.MaskCode(input.Code)), logger.Error(err))
		return nil, err
	}

	gc := giftCardFromRow(row)

	if !gc.IsActive {
		_ = tx.Rollback()
		return nil, domain.ErrGiftCardInactive
	}
	// Expiry is stored as midnight of the expiry date; a card is valid
	// through that entire day, so compare against the end of it.
	if gc.ExpiresAt != nil && !time.Now().Before(gc.ExpiresAt.AddDate(0, 0, 1)) {
		_ = tx.Rollback()
		return nil, domain.ErrGiftCardExpired
	}
	balanceCents, err := decimalToCents(row.CurrentBalance)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if balanceCents < amountCents {
		_ = tx.Rollback()
		return nil, domain.ErrInsufficientBalance
	}

	newBalanceCents := balanceCents - amountCents

	if err := qtx.UpdateGiftCardBalance(ctx, sqlc.UpdateGiftCardBalanceParams{
		CurrentBalance: newBalanceCents.String(),
		ID:             gc.ID,
		BusinessID:     businessID,
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to redeem gift card",
			logger.Int("business_id", int(businessID)), logger.String("code", domain.MaskCode(input.Code)), logger.Error(err))
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO pos_gift_card_transactions
			(gift_card_id, owner_id, business_id, type, order_ref, amount, balance_after, created_by)
		VALUES (?, ?, ?, 'redeem', ?, ?, ?, ?)`,
		gc.ID, gc.OwnerID, businessID, orderRef, amountCents.String(), newBalanceCents.String(), actorID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit redeem transaction",
			logger.Int("business_id", int(businessID)), logger.String("code", domain.MaskCode(input.Code)), logger.Error(err))
		return nil, err
	}

	gc.CurrentBalance = money.CentsToDollars(int64(newBalanceCents))
	return gc, nil
}

func lockPaidOrder(ctx context.Context, tx *sql.Tx, businessID, orderID int64, amount money.Amount) (string, error) {
	var totalString, paymentStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT total, payment_status
		FROM pos_orders WHERE business_id = ? AND id = ? FOR UPDATE`, businessID, orderID).
		Scan(&totalString, &paymentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrInvalidTender
		}
		return "", err
	}
	orderTotal, err := decimalToCents(totalString)
	if err != nil {
		return "", err
	}
	if paymentStatus != "paid" || orderTotal < amount {
		return "", domain.ErrInvalidTender
	}
	orderRef := fmt.Sprintf("order:%d", orderID)
	var issuedString string
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM pos_gift_card_transactions
		WHERE business_id = ? AND type = 'issue' AND order_ref = ?`, businessID, orderRef).
		Scan(&issuedString); err != nil {
		return "", err
	}
	issued, err := decimalToCents(issuedString)
	if err != nil {
		return "", err
	}
	if issued+amount > orderTotal {
		return "", domain.ErrInvalidTender
	}
	return orderRef, nil
}

func lockGiftCardOrder(ctx context.Context, tx *sql.Tx, businessID, orderID int64, amount money.Amount) (string, error) {
	var totalString, status string
	if err := tx.QueryRowContext(ctx, `
		SELECT total, status
		FROM pos_orders WHERE business_id = ? AND id = ? FOR UPDATE`, businessID, orderID).
		Scan(&totalString, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrInvalidOrder
		}
		return "", err
	}
	if status == "cancelled" || status == "refunded" || status == "voided" {
		return "", domain.ErrInvalidOrder
	}
	orderTotal, err := decimalToCents(totalString)
	if err != nil {
		return "", err
	}
	if orderTotal <= 0 || amount > orderTotal {
		return "", domain.ErrInvalidOrder
	}
	orderRef := fmt.Sprintf("order:%d", orderID)
	var redeemedString string
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM pos_gift_card_transactions
		WHERE business_id = ? AND type = 'redeem' AND order_ref = ?`, businessID, orderRef).
		Scan(&redeemedString); err != nil {
		return "", err
	}
	redeemed, err := decimalToCents(redeemedString)
	if err != nil {
		return "", err
	}
	if redeemed+amount > orderTotal {
		return "", domain.ErrInvalidOrder
	}
	return orderRef, nil
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

func decimalToCents(s string) (money.Amount, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	return money.FromString(s)
}

func decimalToFloat(s string) float64 {
	cents, err := decimalToCents(s)
	if err != nil {
		logger.Global().Warn("failed to parse gift card decimal value from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return money.CentsToDollars(int64(cents))
}

func dollarsToCents(dollars float64) (money.Amount, error) {
	if math.IsNaN(dollars) || math.IsInf(dollars, 0) {
		return 0, money.ErrInvalidAmount
	}
	return money.Amount(money.DollarsToCents(dollars)), nil
}
