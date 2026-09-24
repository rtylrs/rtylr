package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/financeaccounts"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	queries         *sqlc.Queries
	pool            *sql.DB
	accountProvider financeaccounts.PayrollAccountProvider
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries:         sqlc.New(pool),
		pool:            pool,
		accountProvider: unavailablePayrollAccountProvider{},
	}
}

func (r *Repository) SetPayrollAccountProvider(provider financeaccounts.PayrollAccountProvider) {
	if provider != nil {
		r.accountProvider = provider
	}
}

func (r *Repository) List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.JournalEntry, error) {
	rows, err := r.queries.ListJournalEntries(ctx, sqlc.ListJournalEntriesParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list journal entries", logger.Error(err))
		return nil, err
	}

	entries := make([]domain.JournalEntry, 0, len(rows))
	entryIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, domain.JournalEntry{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			Date:        row.Date,
			Description: nullStr(row.Description),
			Status:      row.Status,
			PostedAt:    nullTime(row.PostedAt),
			CreatedAt:   nullTime(row.CreatedAt),
		})
		entryIDs = append(entryIDs, row.ID)
	}

	// Batch-load all lines for the listed entries in a single query, then
	// attach them in memory. This replaces a per-entry GetLines call (N+1).
	linesByEntry, err := r.getLinesForEntries(ctx, entryIDs)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Lines = linesByEntry[entries[i].ID]
	}

	return entries, nil
}

// getLinesForEntries loads all journal lines for the given entry IDs in one
// query and groups them by entry ID. An entry with no lines maps to nil.
func (r *Repository) getLinesForEntries(ctx context.Context, entryIDs []int64) (map[int64][]domain.JournalLine, error) {
	out := make(map[int64][]domain.JournalLine, len(entryIDs))
	if len(entryIDs) == 0 {
		return out, nil
	}

	rows, err := r.queries.GetJournalLinesForEntries(ctx, entryIDs)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch-load journal lines", logger.Error(err))
		return nil, err
	}

	for _, row := range rows {
		out[row.EntryID] = append(out[row.EntryID], domain.JournalLine{
			ID:          row.ID,
			EntryID:     row.EntryID,
			AccountID:   row.AccountID,
			AccountName: nullStr(row.AccountName),
			Debit:       parseFloatNull(row.Debit),
			Credit:      parseFloatNull(row.Credit),
		})
	}

	return out, nil
}

func (r *Repository) Get(ctx context.Context, businessID, id int64) (domain.JournalEntry, error) {
	row, err := r.queries.GetJournalEntry(ctx, sqlc.GetJournalEntryParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.JournalEntry{}, domain.ErrJournalEntryNotFound
		}
		logger.Ctx(ctx).Error("Failed to get journal entry", logger.Error(err))
		return domain.JournalEntry{}, err
	}

	entry := domain.JournalEntry{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Date:        row.Date,
		Description: nullStr(row.Description),
		Status:      row.Status,
		PostedAt:    nullTime(row.PostedAt),
		CreatedAt:   nullTime(row.CreatedAt),
	}

	linesByEntry, err := r.getLinesForEntries(ctx, []int64{id})
	if err != nil {
		return domain.JournalEntry{}, err
	}
	entry.Lines = linesByEntry[id]
	return entry, nil
}

func (r *Repository) Post(ctx context.Context, businessID, id int64) error {
	// Verify the entry exists and read its current status up-front so the
	// service can distinguish 404 from 409.
	entry, err := r.Get(ctx, businessID, id)
	if err != nil {
		return err
	}
	if entry.Status == "posted" {
		return domain.ErrJournalEntryAlreadyPosted
	}

	res, err := r.queries.PostJournalEntry(ctx, sqlc.PostJournalEntryParams{ID: id, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to post journal entry",
			logger.Error(err),
			logger.Any("business_id", businessID),
			logger.Any("entry_id", id),
		)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Raced with another caller — entry was posted between the read and
		// the update.
		return domain.ErrJournalEntryAlreadyPosted
	}

	return nil
}

func (r *Repository) PostPayrollJournal(ctx context.Context, input domain.PayrollJournalInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	var entryID int64
	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT id, status
		FROM finance_journal_entries
		WHERE business_id = ? AND source_type = ? AND source_id = ?
		FOR UPDATE`, input.BusinessID, payrollSourceType, input.RunID).Scan(&entryID, &status)
	if err == nil {
		if status == "draft" {
			if _, err := tx.ExecContext(ctx, `
				UPDATE finance_journal_entries
				SET status = 'posted', posted_at = CURRENT_TIMESTAMP
				WHERE id = ? AND business_id = ? AND status = 'draft'`, entryID, input.BusinessID); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return err
	}

	mapping, err := r.accountProvider.GetPayrollAccountMapping(ctx, input.BusinessID)
	if err != nil || mapping.BusinessID != input.BusinessID {
		_ = tx.Rollback()
		if err != nil {
			return err
		}
		return financeaccounts.ErrPayrollAccountsNotConfigured
	}
	expenseAccountID, err := r.lockConfiguredPayrollAccount(ctx, tx, input.BusinessID, mapping.ExpenseAccountID, "expense")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	payableAccountID, err := r.lockConfiguredPayrollAccount(ctx, tx, input.BusinessID, mapping.PayableAccountID, "liability")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	amounts, err := derivePayrollAmounts(input.TotalGross, input.TotalNet)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	gross, net, deductions := amounts.gross, amounts.net, amounts.deductions
	description := fmt.Sprintf("Payroll %d: %s to %s", input.RunID, input.PeriodStart, input.PeriodEnd)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO finance_journal_entries
			(owner_id, business_id, date, description, status, source_type, source_id, posted_at)
		VALUES (?, ?, ?, ?, 'posted', ?, ?, CURRENT_TIMESTAMP)`,
		input.OwnerID, input.BusinessID, input.PeriodEnd, description, payrollSourceType, input.RunID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	entryID, err = result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	lines := []payrollLine{
		{accountID: expenseAccountID, debit: gross},
		{accountID: payableAccountID, credit: net},
	}
	if deductions > 0 {
		deductionsAccountID, accountErr := r.lockConfiguredPayrollAccount(ctx, tx, input.BusinessID, mapping.DeductionsPayableAccountID, "liability")
		if accountErr != nil {
			_ = tx.Rollback()
			return accountErr
		}
		lines = append(lines, payrollLine{accountID: deductionsAccountID, credit: deductions})
	}
	var totalDebits, totalCredits money.Amount
	for _, line := range lines {
		totalDebits += line.debit
		totalCredits += line.credit
	}
	// Exact, because both sides are integers. derivePayrollAmounts already
	// guarantees these three lines close, so this guards a future fourth line
	// added without updating the derivation, not the arithmetic above.
	if totalDebits != totalCredits {
		_ = tx.Rollback()
		return fmt.Errorf("payroll journal is unbalanced: debits=%s credits=%s", totalDebits, totalCredits)
	}
	for _, line := range lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO finance_journal_lines (entry_id, account_id, debit, credit)
			VALUES (?, ?, ?, ?)`, entryID, line.accountID, line.debit.String(), line.credit.String()); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) lockConfiguredPayrollAccount(ctx context.Context, tx *sql.Tx, businessID, accountID int64, accountType string) (int64, error) {
	if accountID <= 0 {
		return 0, financeaccounts.ErrPayrollAccountsNotConfigured
	}
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM finance_accounts
		WHERE id = ? AND business_id = ? AND type = ? AND is_active = TRUE
		FOR UPDATE`, accountID, businessID, accountType).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, financeaccounts.ErrPayrollAccountsNotConfigured
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

type unavailablePayrollAccountProvider struct{}

func (unavailablePayrollAccountProvider) GetPayrollAccountMapping(context.Context, int64) (financeaccounts.PayrollAccountMapping, error) {
	return financeaccounts.PayrollAccountMapping{}, financeaccounts.ErrPayrollAccountsNotConfigured
}

// payrollLine is one ledger line of a payroll entry, in cents.
type payrollLine struct {
	accountID int64
	debit     money.Amount
	credit    money.Amount
}

// payrollAmounts is the balanced set of amounts a payroll entry posts.
type payrollAmounts struct {
	gross      money.Amount
	net        money.Amount
	deductions money.Amount
}

// derivePayrollAmounts turns a payroll export's float64 gross and net into the
// exact cent amounts the ledger stores.
//
// Deliberately pure: no database and no transaction. The double-entry
// arithmetic is the part worth testing exhaustively, and it used to be
// unreachable without a MySQL instance and three configured accounts, which is
// why the posting path had no test at all.
//
// The payroll port carries these as float64, so they are converted once, here,
// and every amount after this point is an integer. The path used to be float64
// throughout with math.Round(value*100)/100 standing in for a cent boundary,
// which made the balance assertion unreliable because it was written in the
// same arithmetic it was guarding (#1183).
func derivePayrollAmounts(grossAmount, netAmount float64) (payrollAmounts, error) {
	// Refused rather than rounded. Rounding a three-decimal amount into the
	// ledger is what the old half-cent tolerance allowed.
	gross, err := money.FromFloat64(grossAmount)
	if err != nil {
		return payrollAmounts{}, fmt.Errorf("payroll gross: %w", err)
	}
	net, err := money.FromFloat64(netAmount)
	if err != nil {
		return payrollAmounts{}, fmt.Errorf("payroll net: %w", err)
	}

	// Derived as a residual, so the three lines close by construction rather
	// than by each being rounded independently.
	deductions := gross - net
	if deductions < 0 {
		return payrollAmounts{}, fmt.Errorf("payroll deductions cannot be negative: gross=%s net=%s", gross, net)
	}
	if gross != net+deductions {
		return payrollAmounts{}, fmt.Errorf("payroll amounts do not balance: gross=%s net=%s deductions=%s", gross, net, deductions)
	}

	return payrollAmounts{gross: gross, net: net, deductions: deductions}, nil
}

const payrollSourceType = "payroll_run"

func (r *Repository) GetLines(ctx context.Context, entryID int64) ([]domain.JournalLine, error) {
	rows, err := r.queries.GetJournalLines(ctx, entryID)
	if err != nil {
		return nil, err
	}

	lines := make([]domain.JournalLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, domain.JournalLine{
			ID:          row.ID,
			EntryID:     row.EntryID,
			AccountID:   row.AccountID,
			AccountName: nullStr(row.AccountName),
			Debit:       parseFloatNull(row.Debit),
			Credit:      parseFloatNull(row.Credit),
		})
	}

	return lines, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.JournalEntryInput) (int64, error) {
	entryDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse journal entry date", logger.Error(err))
		return 0, err
	}

	// A journal entry and its lines are a single double-entry record: persisting
	// the header but only some lines would leave an unbalanced, corrupt entry.
	// Write them all-or-nothing in one transaction.
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

	accountIDs := make([]int64, 0, len(input.Lines))
	seenAccounts := make(map[int64]struct{}, len(input.Lines))
	for _, line := range input.Lines {
		if _, seen := seenAccounts[line.AccountID]; seen {
			continue
		}
		seenAccounts[line.AccountID] = struct{}{}
		accountIDs = append(accountIDs, line.AccountID)
	}
	lockedAccounts, err := qtx.LockJournalAccounts(ctx, sqlc.LockJournalAccountsParams{
		BusinessID: businessID,
		AccountIds: accountIDs,
	})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if len(lockedAccounts) != len(accountIDs) {
		_ = tx.Rollback()
		return 0, fmt.Errorf("%w: every account_id must belong to the active business chart of accounts", domain.ErrInvalidJournalEntry)
	}

	res, err := qtx.CreateJournalEntry(ctx, sqlc.CreateJournalEntryParams{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Date:        entryDate,
		Description: nullableString(input.Description),
	})
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to create journal entry", logger.Error(err))
		return 0, err
	}

	entryID, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	for i, line := range input.Lines {
		// Converted to cents rather than formatted from the float64 with
		// strconv, so the value written is the value the caller meant and the
		// column holds it exactly (#1183). ValidateJournalEntryInput has
		// already refused anything that is not a whole number of cents, so
		// these conversions cannot fail in practice; the errors are handled
		// rather than discarded because a future caller may reach Create
		// without going through the service.
		debit, convErr := money.FromFloat64(line.Debit)
		if convErr != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("line[%d]: debit: %w", i, convErr)
		}
		credit, convErr := money.FromFloat64(line.Credit)
		if convErr != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("line[%d]: credit: %w", i, convErr)
		}

		if err := qtx.CreateJournalLine(ctx, sqlc.CreateJournalLineParams{
			EntryID:   entryID,
			AccountID: line.AccountID,
			Debit:     nullableString(debit.String()),
			Credit:    nullableString(credit.String()),
		}); err != nil {
			_ = tx.Rollback()
			logger.Ctx(ctx).Error("Failed to create journal line", logger.Error(err))
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit journal entry", logger.Error(err))
		return 0, err
	}

	return entryID, nil
}

// CountByStatus returns how many journal entries the owner has for each
// status (e.g. "draft", "posted"). An empty map is returned when the owner
// has no entries.
func (r *Repository) CountByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountJournalEntriesByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count journal entries by status", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.Cnt
	}

	return counts, nil
}

// SumLinesByAccountID returns the total debit and credit amounts posted to a
// given chart-of-accounts ID across all posted journal entries for an owner.
func (r *Repository) SumLinesByAccountID(ctx context.Context, businessID, accountID int64) (debit, credit float64, err error) {
	row, err := r.queries.SumJournalLinesByAccount(ctx, sqlc.SumJournalLinesByAccountParams{
		BusinessID: businessID,
		AccountID:  accountID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum journal lines by account", logger.Error(err))
		return 0, 0, err
	}

	return parseFloatAny(row.TotalDebit), parseFloatAny(row.TotalCredit), nil
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

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse nullable float from database",
			logger.String("value", s.String),
			logger.Error(err),
		)

		return 0
	}

	return v
}

// parseFloatAny converts an aggregate value returned by sqlc as interface{}
// (e.g. COALESCE(SUM(...), 0)) into a float64. Unparseable values are logged
// at Warn level and coerced to 0 rather than discarded silently.
func parseFloatAny(v any) float64 {
	switch val := v.(type) {
	case nil:
		return 0
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return parseFloatString(string(val))
	case string:
		return parseFloatString(val)
	default:
		return parseFloatString(fmt.Sprintf("%v", val))
	}
}

func parseFloatString(s string) float64 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse aggregate float from database",
			logger.String("value", s),
			logger.Error(err),
		)

		return 0
	}

	return v
}
