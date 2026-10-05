package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/journal/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/financeaccounts"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

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
