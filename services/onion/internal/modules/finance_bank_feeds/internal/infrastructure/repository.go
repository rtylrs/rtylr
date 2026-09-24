package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

const listBankFeedTransactionsQuery = `
SELECT id, owner_id, business_id, bank_account_id, amount_minor_units, currency, occurred_at,
       description, external_id, matched_expense_id, matched_invoice_id, created_at
FROM finance_bank_feeds
WHERE business_id = ? AND (? = 0 OR bank_account_id = ?)
ORDER BY occurred_at DESC, id DESC
LIMIT ? OFFSET ?`

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	return r.ListByAccount(ctx, businessID, 0, limit, offset)
}

func (r *Repository) ListByAccount(ctx context.Context, businessID, bankAccountID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	rows, err := r.pool.QueryContext(ctx, listBankFeedTransactionsQuery,
		businessID, bankAccountID, bankAccountID, limit, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bank feed transactions",
			logger.Error(err),
			logger.Any("business_id", businessID),
			logger.Any("bank_account_id", bankAccountID),
		)
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.BankFeedTransaction, 0)
	for rows.Next() {
		var row sqlc.FinanceBankFeed
		if err := rows.Scan(
			&row.ID,
			&row.OwnerID,
			&row.BusinessID,
			&row.BankAccountID,
			&row.AmountMinorUnits,
			&row.Currency,
			&row.OccurredAt,
			&row.Description,
			&row.ExternalID,
			&row.MatchedExpenseID,
			&row.MatchedInvoiceID,
			&row.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, bankFeedTransactionFromRow(row))
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	if bankAccountID == 0 {
		total, err = r.queries.CountBankFeedTransactions(ctx, businessID)
	} else {
		err = r.pool.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM finance_bank_feeds
			WHERE business_id = ? AND bank_account_id = ?`, businessID, bankAccountID).Scan(&total)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count bank feed transactions",
			logger.Error(err),
			logger.Any("business_id", businessID),
			logger.Any("bank_account_id", bankAccountID),
		)
		return nil, 0, err
	}

	return items, total, nil
}

func (r *Repository) BankAccountExists(ctx context.Context, businessID, bankAccountID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM finance_bank_accounts
			WHERE id = ? AND business_id = ?
		)`, bankAccountID, businessID).Scan(&exists)
	return exists, err
}

func bankFeedTransactionFromRow(row sqlc.FinanceBankFeed) domain.BankFeedTransaction {
	return domain.BankFeedTransaction{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		BusinessID:       row.BusinessID,
		BankAccountID:    row.BankAccountID,
		AmountMinorUnits: row.AmountMinorUnits,
		Currency:         row.Currency,
		OccurredAt:       row.OccurredAt,
		Description:      nullStr(row.Description),
		ExternalID:       row.ExternalID,
		MatchedExpenseID: nullInt64Ptr(row.MatchedExpenseID),
		MatchedInvoiceID: nullInt64Ptr(row.MatchedInvoiceID),
		CreatedAt:        nullTime(row.CreatedAt),
	}
}

// SumAmountByAccountID returns the net amount (sum of AmountMinorUnits) for
// all feed transactions belonging to a specific bank account. Positive values
// indicate net credits; negative values indicate net debits.
func (r *Repository) SumAmountByAccountID(ctx context.Context, businessID, bankAccountID int64) (int64, error) {
	result, err := r.queries.SumBankFeedAmountByAccount(ctx, sqlc.SumBankFeedAmountByAccountParams{
		BusinessID:    businessID,
		BankAccountID: bankAccountID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank feed amounts by account", logger.Error(err))
		return 0, err
	}

	return toInt64(result), nil
}

// CountUnmatched returns the number of transactions that have neither a
// matched_expense_id nor a matched_invoice_id.
func (r *Repository) CountUnmatched(ctx context.Context, businessID int64) (int64, error) {
	cnt, err := r.queries.CountUnmatchedBankFeedTransactions(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count unmatched bank feed transactions", logger.Error(err))
		return 0, err
	}

	return cnt, nil
}

// Sync applies imported feed rows to connected account balances exactly once.
// The account's last_synced_at is the cursor, and both the balance update and
// cursor advance are committed in one transaction so retries are idempotent.
func (r *Repository) Sync(ctx context.Context, businessID int64) (int, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)
	accounts, err := qtx.ListSyncBankAccounts(ctx, businessID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	imported := int64(0)
	for _, account := range accounts {
		var cursor *time.Time
		if account.LastSyncedAt.Valid {
			cursor = &account.LastSyncedAt.Time
		}
		amountTotal, count, err := sumBankFeedAmounts(ctx, tx, businessID, account.ID, cursor)
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}

		balance := parseFloatNull(account.Balance) + float64(amountTotal)/100
		if _, err := qtx.UpdateSyncedBankAccount(ctx, sqlc.UpdateSyncedBankAccountParams{
			Balance:    nullableString(fmtFloat(balance)),
			ID:         account.ID,
			BusinessID: businessID,
		}); err != nil {
			_ = tx.Rollback()
			return 0, err
		}
		imported += count
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(imported), nil // #nosec G115 -- imported row count is bounded by the database and the API uses int.
}

func sumBankFeedAmounts(ctx context.Context, tx *sql.Tx, businessID, bankAccountID int64, cursor *time.Time) (int64, int64, error) {
	query := `
		SELECT COALESCE(SUM(amount_minor_units), 0), COUNT(*)
		FROM finance_bank_feeds
		WHERE business_id = ? AND bank_account_id = ?`
	args := []any{businessID, bankAccountID}
	if cursor != nil {
		query += " AND created_at > ?"
		args = append(args, *cursor)
	}

	var rawTotal interface{}
	var count int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&rawTotal, &count); err != nil {
		return 0, 0, err
	}
	return toInt64(rawTotal), count, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid {
		return 0
	}
	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse synced bank account balance",
			logger.String("value", s.String),
			logger.Error(err),
		)
		return 0
	}
	return v
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func nullInt64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

// toInt64 normalizes the result of COALESCE(SUM(...), 0). MySQL returns the
// sum of a BIGINT column as DECIMAL, which the driver may surface as int64,
// []byte, or string depending on configuration.
func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case []byte:
		return parseInt64Str(string(val))
	case string:
		return parseInt64Str(val)
	default:
		return parseInt64Str(fmt.Sprintf("%v", val))
	}
}

func parseInt64Str(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// A DECIMAL sum may carry a fractional/scientific suffix; fall back
		// to float parsing before giving up so the magnitude is preserved.
		if f, ferr := strconv.ParseFloat(s, 64); ferr == nil {
			return int64(f)
		}
		logger.Global().Warn("failed to parse bank feed amount sum from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return v
}
