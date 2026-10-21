package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/journal/internal/infrastructure/persistence/sqlc"
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
	if len(entry.Lines) < 2 {
		return fmt.Errorf("%w: a journal entry must have at least two lines", domain.ErrInvalidJournalEntry)
	}

	var totalDebit, totalCredit money.Amount
	for i, line := range entry.Lines {
		if line.Debit > 0 && line.Credit > 0 {
			return fmt.Errorf("%w: line[%d]: cannot have both debit and credit", domain.ErrInvalidJournalEntry, i)
		}
		debit, convErr := money.FromFloat64(line.Debit)
		if convErr != nil {
			return fmt.Errorf("%w: line[%d]: debit: %w", domain.ErrInvalidJournalEntry, i, convErr)
		}
		credit, convErr := money.FromFloat64(line.Credit)
		if convErr != nil {
			return fmt.Errorf("%w: line[%d]: credit: %w", domain.ErrInvalidJournalEntry, i, convErr)
		}
		totalDebit += debit
		totalCredit += credit
	}
	if totalDebit != totalCredit || totalDebit <= 0 {
		return fmt.Errorf("%w: journal entry is not balanced: total_debit=%s total_credit=%s", domain.ErrInvalidJournalEntry, totalDebit, totalCredit)
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
