package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	reportingCurrency func(context.Context, int64) (string, error)
	pool              *sql.DB
	queries           *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.BankAccount, error) {
	accounts, _, err := r.ListPage(ctx, businessID, 1, 500, "name", "asc")
	return accounts, err
}

func (r *Repository) ListPage(ctx context.Context, businessID int64, page, limit int32, sortBy, sortDir string) ([]domain.BankAccount, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	offset := int64(page-1) * int64(limit)
	sortColumn := bankAccountSortColumn(sortBy)
	direction := "ASC"
	if strings.EqualFold(strings.TrimSpace(sortDir), "desc") {
		direction = "DESC"
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM finance_bank_accounts WHERE business_id = ?`, businessID).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count bank accounts", logger.Error(err))
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, owner_id, business_id, name, institution, account_type, last_four,
		       balance, currency, is_primary, is_connected, last_synced_at, created_at
		FROM finance_bank_accounts
		WHERE business_id = ?
		ORDER BY %s %s, id %s
		LIMIT ? OFFSET ?`, sortColumn, direction, direction)
	rows, err := r.pool.QueryContext(ctx, query, businessID, limit, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bank accounts", logger.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.BankAccount, 0, limit)
	for rows.Next() {
		account, scanErr := scanBankAccount(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, account)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.BankAccount, error) {
	row := r.pool.QueryRowContext(ctx, `
		SELECT id, owner_id, business_id, name, institution, account_type, last_four,
		       balance, currency, is_primary, is_connected, last_synced_at, created_at
		FROM finance_bank_accounts
		WHERE business_id = ? AND id = ?`, businessID, id)
	account, err := scanBankAccount(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.BankAccount{}, domain.ErrBankAccountNotFound
	}
	if err != nil {
		return domain.BankAccount{}, err
	}
	return account, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.BankAccountInput) (int64, error) {
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "USD"
	}

	res, err := r.pool.ExecContext(ctx, `
		INSERT INTO finance_bank_accounts
		(owner_id, business_id, name, institution, account_type, last_four, balance, currency, is_primary, is_connected)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		businessID, businessID, input.Name, nullableString(input.Institution),
		nullableString(input.AccountType), nullableString(input.LastFour),
		nullableString(input.Balance.String()), nullableString(currency),
		input.IsPrimary, input.IsConnected,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create bank account", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

// TotalBalance returns the sum of balances in the business reporting currency.
func (r *Repository) TotalBalance(ctx context.Context, businessID int64) (money.Amount, error) {
	if r.reportingCurrency == nil {
		return 0, errors.New("bank account reporting currency reader is not configured")
	}
	currency, err := r.reportingCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var result any
	err = r.pool.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CAST(COALESCE(a.balance, 0) AS DECIMAL(30,2))), 0)
		FROM finance_bank_accounts a
		WHERE a.business_id = ?
		  AND COALESCE(NULLIF(UPPER(TRIM(a.currency)), ''), 'USD') =
		      ?`, businessID, currency).Scan(&result)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank account balances", logger.Error(err))
		return 0, err
	}

	return toMoney(result), nil
}

// CountConnected returns the number of bank accounts with is_connected = TRUE.
func (r *Repository) CountConnected(ctx context.Context, businessID int64) (int64, error) {
	cnt, err := r.queries.CountConnectedBankAccounts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count connected bank accounts", logger.Error(err))
		return 0, err
	}

	return cnt, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

type bankAccountScanner interface {
	Scan(dest ...any) error
}

func scanBankAccount(scanner bankAccountScanner) (domain.BankAccount, error) {
	var (
		account     domain.BankAccount
		institution sql.NullString
		accountType sql.NullString
		lastFour    sql.NullString
		balance     sql.NullString
		currency    sql.NullString
		isPrimary   sql.NullBool
		isConnected sql.NullBool
		lastSynced  sql.NullTime
		createdAt   sql.NullTime
	)
	if err := scanner.Scan(
		&account.ID,
		&account.OwnerID,
		&account.BusinessID,
		&account.Name,
		&institution,
		&accountType,
		&lastFour,
		&balance,
		&currency,
		&isPrimary,
		&isConnected,
		&lastSynced,
		&createdAt,
	); err != nil {
		return domain.BankAccount{}, err
	}
	account.Institution = nullStr(institution)
	account.AccountType = nullStr(accountType)
	account.LastFour = nullStr(lastFour)
	account.Balance = parseMoneyNull(balance)
	account.Currency = nullStr(currency)
	account.IsPrimary = nullBool(isPrimary)
	account.IsConnected = nullBool(isConnected)
	account.LastSyncedAt = nullTime(lastSynced)
	account.CreatedAt = nullTime(createdAt)
	return account, nil
}

func bankAccountSortColumn(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "balance":
		return "balance"
	case "currency":
		return "currency"
	case "created_at", "createdat":
		return "created_at"
	case "institution", "bank_name":
		return "institution"
	case "account_type":
		return "account_type"
	default:
		return "name"
	}
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

func parseMoneyNull(s sql.NullString) money.Amount {
	if !s.Valid {
		return 0
	}
	value, err := money.FromString(s.String)
	if err != nil {
		logger.Global().Warn("failed to parse bank account balance from database",
			logger.String("value", s.String),
			logger.Error(err),
		)
		return 0
	}
	return value
}

func toMoney(v interface{}) money.Amount {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case money.Amount:
		return val
	case float64:
		parsed, err := money.FromFloat64(val)
		if err == nil {
			return parsed
		}
	case float32:
		parsed, err := money.FromFloat64(float64(val))
		if err == nil {
			return parsed
		}
	case int64:
		parsed, err := money.FromFloat64(float64(val))
		if err == nil {
			return parsed
		}
	case []byte:
		return parseMoneyString(string(val))
	case string:
		return parseMoneyString(val)
	default:
		return parseMoneyString(fmt.Sprintf("%v", val))
	}
	return 0
}

func parseMoneyString(s string) money.Amount {
	if s == "" {
		return 0
	}
	value, err := money.FromString(s)
	if err != nil {
		logger.Global().Warn("failed to parse bank account balance sum from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return value
}

func (r *Repository) WithReportingCurrency(read func(context.Context, int64) (string, error)) *Repository {
	r.reportingCurrency = read
	return r
}
