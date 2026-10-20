package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

// ListUnpostedSources returns the business's approved expenses and invoice
// payments that have no journal entry yet, oldest first.
func (r *Repository) ListUnpostedSources(ctx context.Context, businessID int64) ([]domain.OperationalSource, error) {
	var out []domain.OperationalSource

	// cross-schema read from expenses — approved expenses post to the journal.
	expenses, err := r.pool.QueryContext(ctx, `
		SELECT e.id, e.owner_id, e.expense_date, e.description, e.amount,
		       COALESCE(NULLIF(e.currency, ''), 'USD'), COALESCE(e.payment_method, 'cash')
		FROM finance_expenses e
		LEFT JOIN finance_journal_entries j
		       ON j.business_id = e.business_id AND j.source_type = ? AND j.source_id = e.id
		WHERE e.business_id = ? AND e.status = 'approved' AND j.id IS NULL
		ORDER BY e.expense_date ASC, e.id ASC`, domain.SourceExpense, businessID)
	if err != nil {
		return nil, err
	}
	for expenses.Next() {
		var (
			src         = domain.OperationalSource{Kind: domain.SourceExpense, BusinessID: businessID}
			date        time.Time
			description string
			amount      string
			method      string
		)
		if err := expenses.Scan(&src.ID, &src.OwnerID, &date, &description, &amount, &src.Currency, &method); err != nil {
			_ = expenses.Close()
			return nil, err
		}
		if src.Amount, err = money.FromString(amount); err != nil {
			_ = expenses.Close()
			return nil, fmt.Errorf("expense %d amount %q: %w", src.ID, amount, err)
		}
		src.Date = date.Format("2006-01-02")
		src.Description = "Expense: " + description
		src.PaidInCash = method == "cash"
		out = append(out, src)
	}
	if err := expenses.Close(); err != nil {
		return nil, err
	}

	// cross-schema read from invoices — recorded payments post to the journal.
	payments, err := r.pool.QueryContext(ctx, `
		SELECT p.id, p.owner_id, p.payment_date, i.invoice_number, i.client_name, p.amount,
		       COALESCE(NULLIF(p.currency, ''), 'USD'), p.method
		FROM finance_payments p
		JOIN finance_invoices i ON i.id = p.invoice_id AND i.business_id = p.business_id
		LEFT JOIN finance_journal_entries j
		       ON j.business_id = p.business_id AND j.source_type = ? AND j.source_id = p.id
		WHERE p.business_id = ? AND j.id IS NULL
		ORDER BY p.payment_date ASC, p.id ASC`, domain.SourceInvoicePayment, businessID)
	if err != nil {
		return nil, err
	}
	defer payments.Close()
	for payments.Next() {
		var (
			src           = domain.OperationalSource{Kind: domain.SourceInvoicePayment, BusinessID: businessID}
			date          time.Time
			invoiceNumber string
			client        string
			amount        string
			method        string
		)
		if err := payments.Scan(&src.ID, &src.OwnerID, &date, &invoiceNumber, &client, &amount, &src.Currency, &method); err != nil {
			return nil, err
		}
		if src.Amount, err = money.FromString(amount); err != nil {
			return nil, fmt.Errorf("payment %d amount %q: %w", src.ID, amount, err)
		}
		src.Date = date.Format("2006-01-02")
		src.Description = fmt.Sprintf("Payment for %s · %s", invoiceNumber, client)
		src.PaidInCash = method == "cash"
		out = append(out, src)
	}
	return out, payments.Err()
}

// PostOperationalEntry writes one posted, balanced two-line entry for the
// source. It reports false when the source was already posted: the unique key
// on (business, source type, source id) turns a concurrent second post into a
// duplicate-key error rather than a second entry.
func (r *Repository) PostOperationalEntry(ctx context.Context, src domain.OperationalSource, debitAccountID, creditAccountID int64) (bool, error) {
	if src.Amount <= 0 {
		return false, nil
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO finance_journal_entries
			(owner_id, business_id, date, description, status, source_type, source_id, posted_at)
		VALUES (?, ?, ?, ?, 'posted', ?, ?, CURRENT_TIMESTAMP)`,
		src.OwnerID, src.BusinessID, src.Date, src.Description, src.Kind, src.ID)
	if err != nil {
		_ = tx.Rollback()
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return false, nil
		}
		return false, err
	}
	entryID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	for _, line := range []struct {
		accountID     int64
		debit, credit money.Amount
	}{
		{accountID: debitAccountID, debit: src.Amount},
		{accountID: creditAccountID, credit: src.Amount},
	} {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO finance_journal_lines (entry_id, account_id, debit, credit)
			VALUES (?, ?, ?, ?)`, entryID, line.accountID, line.debit.String(), line.credit.String()); err != nil {
			_ = tx.Rollback()
			return false, err
		}
	}
	return true, tx.Commit()
}
