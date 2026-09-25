package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func updateCustomer(ctx context.Context, q *sqlc.Queries, input domain.CustomerUpdateInput) error {
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	result, err := q.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
		FirstName:   input.FirstName,
		LastName:    sql.NullString{String: input.LastName, Valid: input.LastName != ""},
		Email:       sql.NullString{String: input.Email, Valid: input.Email != ""},
		Phone:       sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		DateOfBirth: nullableTime(input.DateOfBirth),
		Gender:      sqlc.NullCrmCustomersGender{CrmCustomersGender: sqlc.CrmCustomersGender(input.Gender), Valid: input.Gender != ""},
		Address:     sql.NullString{String: input.Address, Valid: input.Address != ""},
		City:        sql.NullString{String: input.City, Valid: input.City != ""},
		Country:     sql.NullString{String: input.Country, Valid: input.Country != ""},
		Tags:        tagsJSON,
		Notes:       sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		Consent:     input.Consent,
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update customer", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *Repository) DeleteCustomer(ctx context.Context, businessID, customerID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := r.queries.WithTx(tx).DeleteCustomer(ctx, sqlc.DeleteCustomerParams{
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete customer", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCustomerNotFound
	}

	// The CRM schema has no foreign keys: rows in these tables would keep
	// referencing the deleted customer forever (campaign/feedback stats join
	// on them). Tables with NOT NULL customer_id are purged; nullable
	// references are detached.
	cleanup := []string{
		"DELETE FROM crm_campaign_sends WHERE business_id = ? AND customer_id = ?",
		"DELETE FROM crm_feedback WHERE business_id = ? AND customer_id = ?",
		"DELETE FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?",
		"DELETE FROM crm_customer_timeline WHERE business_id = ? AND customer_id = ?",
		"UPDATE crm_contacts SET customer_id = NULL WHERE business_id = ? AND customer_id = ?",
	}
	for _, stmt := range cleanup {
		if _, err := tx.ExecContext(ctx, stmt, businessID, customerID); err != nil {
			logger.Ctx(ctx).Error("Failed to clean up customer dependencies", logger.Error(err))

			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateCustomerLoyalty(ctx context.Context, businessID, customerID int64, newPoints int, newTier string) error {
	if newPoints < 0 || int64(newPoints) > domain.MaxLoyaltyPoints {
		return domain.ErrLoyaltyPointsOutOfRange
	}

	err := r.queries.UpdateCustomerLoyalty(ctx, sqlc.UpdateCustomerLoyaltyParams{
		LoyaltyPoints: sql.NullInt32{Int32: int32(newPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		LoyaltyTier: sqlc.NullCrmCustomersLoyaltyTier{
			CrmCustomersLoyaltyTier: sqlc.CrmCustomersLoyaltyTier(newTier),
			Valid:                   newTier != "",
		},
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update customer loyalty", logger.Error(err))
	}
	return err
}

func (r *Repository) RecordVisit(ctx context.Context, businessID, customerID int64, amount float64) error {
	err := r.queries.RecordVisit(ctx, sqlc.RecordVisitParams{
		TotalSpent: sql.NullString{String: strconv.FormatFloat(amount, 'f', 2, 64), Valid: true},
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record visit", logger.Error(err))
	}

	return err
}

// ─── Loyalty Transactions ─────────────────────────────────────────────────────

func (r *Repository) CreateLoyaltyTransaction(ctx context.Context, tx domain.LoyaltyTransaction) error {
	if int64(tx.Points) < -domain.MaxLoyaltyPoints || int64(tx.Points) > domain.MaxLoyaltyPoints ||
		tx.BalanceBefore < 0 || int64(tx.BalanceBefore) > domain.MaxLoyaltyPoints ||
		tx.BalanceAfter < 0 || int64(tx.BalanceAfter) > domain.MaxLoyaltyPoints {
		return domain.ErrLoyaltyPointsOutOfRange
	}

	err := r.queries.CreateLoyaltyTransaction(ctx, sqlc.CreateLoyaltyTransactionParams{
		OwnerID:       tx.OwnerID,
		BusinessID:    tx.BusinessID,
		CustomerID:    tx.CustomerID,
		Type:          sqlc.CrmLoyaltyTransactionsType(tx.Type),
		Points:        int32(tx.Points),        // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BalanceBefore: int32(tx.BalanceBefore), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BalanceAfter:  int32(tx.BalanceAfter),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Description:   sql.NullString{String: tx.Description, Valid: tx.Description != ""},
		ReferenceID:   nullInt64FromPtr(tx.ReferenceID),
		ReferenceType: sql.NullString{String: tx.ReferenceType, Valid: tx.ReferenceType != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
	}

	return err
}

// ApplyLoyaltyDelta serializes the balance update and audit insert in one
// transaction. The customer row lock prevents concurrent awards/redeems from
// calculating from the same stale balance. onCommitted runs only after the
// transaction commits, so callers can invalidate caches without racing a
// rollback.
func (r *Repository) ApplyLoyaltyDelta(ctx context.Context, input domain.LoyaltyTransaction, delta int, tierFor func(int) string, onCommitted func()) error {
	if err := domain.ValidateLoyaltyPoints(int64(input.Points)); err != nil {
		return err
	}
	if delta == 0 || int64(delta) > domain.MaxLoyaltyPoints || int64(delta) < -domain.MaxLoyaltyPoints {
		return domain.ErrLoyaltyPointsOutOfRange
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var balanceBefore int
	err = tx.QueryRowContext(ctx,
		"SELECT loyalty_points FROM crm_customers WHERE id = ? AND business_id = ? FOR UPDATE",
		input.CustomerID, input.BusinessID,
	).Scan(&balanceBefore)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrCustomerNotFound
	}
	if err != nil {
		return err
	}
	if balanceBefore < 0 || int64(balanceBefore) > domain.MaxLoyaltyPoints {
		return domain.ErrLoyaltyBalanceCapacity
	}

	// Check the cumulative result while the row lock is held. Checking only
	// the request amount, or calculating after releasing the lock, lets two
	// valid awards race past the signed INT capacity.
	if delta > 0 && int64(balanceBefore) > domain.MaxLoyaltyPoints-int64(delta) {
		return domain.ErrLoyaltyBalanceCapacity
	}
	if delta < 0 && int64(balanceBefore) < -int64(delta) {
		return domain.ErrInsufficientPoints
	}
	balanceAfter := balanceBefore + delta
	if balanceAfter < 0 {
		return domain.ErrInsufficientPoints
	}
	if int64(balanceAfter) > domain.MaxLoyaltyPoints {
		return domain.ErrLoyaltyBalanceCapacity
	}
	if _, err = tx.ExecContext(ctx,
		"UPDATE crm_customers SET loyalty_points = ?, loyalty_tier = ? WHERE id = ? AND business_id = ?",
		balanceAfter, tierFor(balanceAfter), input.CustomerID, input.BusinessID,
	); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO crm_loyalty_transactions
		  (owner_id, business_id, customer_id, type, points, balance_before, balance_after,
		   description, reference_id, reference_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.BusinessID, input.CustomerID, input.Type, input.Points,
		balanceBefore, balanceAfter, input.Description, input.ReferenceID, input.ReferenceType,
	); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if onCommitted != nil {
		onCommitted()
	}
	return nil
}

func (r *Repository) ListLoyaltyTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var err error
	var total int64
	if customerID == 0 {
		total, err = r.queries.CountAllLoyaltyTransactions(ctx, businessID)
	} else {
		total, err = r.queries.CountLoyaltyTransactions(ctx, sqlc.CountLoyaltyTransactionsParams{
			BusinessID: businessID,
			CustomerID: customerID,
		})
	}
	if err != nil {
		return nil, 0, err
	}

	var rows []sqlc.CrmLoyaltyTransaction
	if customerID == 0 {
		rows, err = r.queries.ListAllLoyaltyTransactions(ctx, sqlc.ListAllLoyaltyTransactionsParams{
			BusinessID: businessID,
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	} else {
		rows, err = r.queries.ListLoyaltyTransactions(ctx, sqlc.ListLoyaltyTransactionsParams{
			BusinessID: businessID,
			CustomerID: customerID,
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))

		return nil, 0, err
	}

	txs := make([]domain.LoyaltyTransaction, 0, len(rows))

	for _, row := range rows {
		tx := domain.LoyaltyTransaction{
			ID:            row.ID,
			OwnerID:       row.OwnerID,
			BusinessID:    row.BusinessID,
			CustomerID:    row.CustomerID,
			Type:          string(row.Type),
			Points:        int(row.Points),
			BalanceBefore: int(row.BalanceBefore),
			BalanceAfter:  int(row.BalanceAfter),
			Description:   row.Description.String,
			ReferenceID:   nullInt64Ptr(row.ReferenceID),
			ReferenceType: row.ReferenceType.String,
			CreatedAt:     row.CreatedAt.Time,
		}
		txs = append(txs, tx)
	}

	return txs, int(total), nil
}

// ─── Loyalty Config ───────────────────────────────────────────────────────────
