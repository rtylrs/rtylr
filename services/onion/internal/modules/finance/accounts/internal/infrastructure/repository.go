package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance/accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/accounts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/financeaccounts"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func (r *Repository) GetPayrollAccountMapping(ctx context.Context, businessID int64) (financeaccounts.PayrollAccountMapping, error) {
	row, err := r.queries.GetPayrollAccountMapping(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return financeaccounts.PayrollAccountMapping{}, financeaccounts.ErrPayrollAccountsNotConfigured
		}
		return financeaccounts.PayrollAccountMapping{}, err
	}

	mapping := financeaccounts.PayrollAccountMapping{
		BusinessID:       row.BusinessID,
		ExpenseAccountID: row.ExpenseAccountID,
		PayableAccountID: row.PayableAccountID,
	}
	if row.DeductionsPayableAccountID.Valid {
		mapping.DeductionsPayableAccountID = row.DeductionsPayableAccountID.Int64
	}
	return mapping, nil
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64, limit, offset int) ([]domain.Account, error) {
	rows, err := r.queries.ListAccounts(ctx, sqlc.ListAccountsParams{BusinessID: businessID, Limit: int32(limit), Offset: int32(offset)}) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list accounts", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Account{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Code:       nullStr(row.Code),
			Type:       string(row.Type),
			Currency:   nullStr(row.Currency),
			Balance:    parseFloatNull(row.Balance),
			IsActive:   nullBool(row.IsActive),
			CreatedAt:  nullTime(row.CreatedAt),
		})
		account := &items[len(items)-1]
		netDebit, _ := strconv.ParseFloat(row.PostedNetDebit, 64)
		if account.Type == "asset" || account.Type == "expense" {
			account.CurrentBalance = account.Balance + netDebit
		} else {
			account.CurrentBalance = account.Balance - netDebit
		}
	}

	return items, nil
}

// EnsureLedgerAccount returns the business's account with this code, type and
// currency, creating it with a zero opening balance when there is none.
func (r *Repository) EnsureLedgerAccount(ctx context.Context, ownerID, businessID int64, code, name, accountType, currency string) (int64, bool, error) {
	id, err := r.queries.FindLedgerAccount(ctx, sqlc.FindLedgerAccountParams{
		BusinessID:  businessID,
		Code:        nullableString(code),
		AccountType: sqlc.FinanceAccountsType(accountType),
		Currency:    nullableString(currency),
	})
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}
	res, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Name:       name,
		Code:       nullableString(code),
		Type:       sqlc.FinanceAccountsType(accountType),
		Currency:   nullableString(currency),
		Balance:    nullableString("0.00"),
		IsActive:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create ledger account", logger.String("code", code), logger.Error(err))
		return 0, false, err
	}
	id, err = res.LastInsertId()
	return id, true, err
}

func (r *Repository) ListForTree(ctx context.Context, businessID int64) ([]domain.Account, error) {
	rows, err := r.queries.ListAccountsForTree(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list accounts for tree", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		parentID := int64(0)
		if row.ParentID.Valid {
			parentID = row.ParentID.Int64
		}
		items = append(items, domain.Account{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Code:       nullStr(row.Code),
			ParentID:   parentID,
			Type:       string(row.Type),
			Currency:   nullStr(row.Currency),
			Balance:    parseFloatNull(row.Balance),
			IsActive:   nullBool(row.IsActive),
			CreatedAt:  nullTime(row.CreatedAt),
		})
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.AccountInput) (int64, error) {
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	res, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       input.Name,
		Code:       nullableString(input.Code),
		Type:       sqlc.FinanceAccountsType(input.Type),
		Currency:   nullableString(currency),
		Balance:    nullableString(fmtFloat(input.Balance)),
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create account", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.AccountInput) error {
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}
	res, err := r.queries.UpdateAccount(ctx, sqlc.UpdateAccountParams{
		Name:       input.Name,
		Code:       nullableString(input.Code),
		Type:       sqlc.FinanceAccountsType(input.Type),
		Currency:   nullableString(currency),
		Balance:    nullableString(fmtFloat(input.Balance)),
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update account", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	res, err := r.queries.DeleteAccount(ctx, sqlc.DeleteAccountParams{ID: id, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete account", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal from database", logger.String("value", s.String), logger.Error(err))
		return 0
	}
	return v
}
