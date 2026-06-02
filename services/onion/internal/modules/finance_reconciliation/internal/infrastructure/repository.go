package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) ListUnmatchedBankFeeds(ctx context.Context, ownerID int64) ([]domain.UnmatchedBankTransaction, error) {
	rows, err := r.queries.ListUnmatchedBankFeeds(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list unmatched bank feeds",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)
		return nil, err
	}

	items := make([]domain.UnmatchedBankTransaction, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.UnmatchedBankTransaction{
			ID:               row.ID,
			BankAccountID:    row.BankAccountID,
			AmountMinorUnits: row.AmountMinorUnits,
			Currency:         row.Currency,
			OccurredAt:       row.OccurredAt,
			Description:      nullStr(row.Description),
			ExternalID:       row.ExternalID,
		})
	}
	return items, nil
}

// ListUnmatchedInvoices reads from finance_invoices. The invoices module owns
// that table; we read directly because reconciliation needs both sides of the
// match and routing this through an in-process RPC would add latency and
// failure modes without meaningful isolation.
func (r *Repository) ListUnmatchedInvoices(ctx context.Context, ownerID int64) ([]domain.UnmatchedInvoice, error) {
	// cross-schema read from finance_invoices — reconciliation needs the outstanding invoice side of the match (owned by the invoices module).
	const q = `
		SELECT i.id, i.invoice_number, i.client_name, i.total, i.currency, i.issue_date, i.due_date
		FROM finance_invoices i
		WHERE i.owner_id = ?
		  AND i.status IN ('draft','sent','viewed','partial','overdue')
		  AND NOT EXISTS (
		      SELECT 1 FROM finance_reconciliation_matches m
		      WHERE m.owner_id = i.owner_id
		        AND m.target_type = 'invoice'
		        AND m.target_id = i.id
		  )
		ORDER BY i.due_date ASC, i.id ASC
		LIMIT 200
	`
	rows, err := r.pool.QueryContext(ctx, q, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list unmatched invoices",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("Failed to close unmatched invoices rows", logger.Error(closeErr))
		}
	}()

	items := make([]domain.UnmatchedInvoice, 0)
	for rows.Next() {
		var (
			id            int64
			invoiceNumber string
			clientName    string
			totalStr      sql.NullString
			currency      sql.NullString
			issueDate     time.Time
			dueDate       time.Time
		)
		if scanErr := rows.Scan(&id, &invoiceNumber, &clientName, &totalStr, &currency, &issueDate, &dueDate); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, domain.UnmatchedInvoice{
			ID:               id,
			InvoiceNumber:    invoiceNumber,
			ClientName:       clientName,
			AmountMinorUnits: decimalStrToMinorUnits(ctx, totalStr),
			Currency:         nullStr(currency),
			IssueDate:        issueDate,
			DueDate:          dueDate,
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return items, nil
}

// ListUnmatchedExpenses reads from finance_expenses. Same cross-schema
// rationale as ListUnmatchedInvoices. The expenses module owns the table but
// reconciliation legitimately needs to see the payable side for match
// suggestions.
func (r *Repository) ListUnmatchedExpenses(ctx context.Context, ownerID int64) ([]domain.UnmatchedInvoice, error) {
	// cross-schema read from finance_expenses — reconciliation needs the outstanding expense side of the match (owned by the expenses module).
	const q = `
		SELECT e.id, e.description, COALESCE(e.vendor, ''), e.amount, e.currency, e.expense_date
		FROM finance_expenses e
		WHERE e.owner_id = ?
		  AND NOT EXISTS (
		      SELECT 1 FROM finance_reconciliation_matches m
		      WHERE m.owner_id = e.owner_id
		        AND m.target_type = 'expense'
		        AND m.target_id = e.id
		  )
		ORDER BY e.expense_date ASC, e.id ASC
		LIMIT 200
	`
	rows, err := r.pool.QueryContext(ctx, q, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list unmatched expenses",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("Failed to close unmatched expenses rows", logger.Error(closeErr))
		}
	}()

	items := make([]domain.UnmatchedInvoice, 0)
	for rows.Next() {
		var (
			id          int64
			description string
			vendor      string
			amountStr   sql.NullString
			currency    sql.NullString
			expenseDate time.Time
		)
		if scanErr := rows.Scan(&id, &description, &vendor, &amountStr, &currency, &expenseDate); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, domain.UnmatchedInvoice{
			ID:               id,
			InvoiceNumber:    description,
			ClientName:       vendor,
			AmountMinorUnits: decimalStrToMinorUnits(ctx, amountStr),
			Currency:         nullStr(currency),
			IssueDate:        expenseDate,
			DueDate:          expenseDate,
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return items, nil
}

func (r *Repository) BankFeedExists(ctx context.Context, ownerID, id int64) (bool, error) {
	_, err := r.queries.GetBankFeedForOwner(ctx, sqlc.GetBankFeedForOwnerParams{ID: id, OwnerID: ownerID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CreateMatch(ctx context.Context, ownerID int64, input domain.MatchInput) (int64, error) {
	matchedBy := sql.NullInt64{}
	if input.MatchedBy > 0 {
		matchedBy = sql.NullInt64{Int64: input.MatchedBy, Valid: true}
	}
	res, err := r.queries.CreateReconciliationMatch(ctx, sqlc.CreateReconciliationMatchParams{
		OwnerID:    ownerID,
		BankFeedID: input.BankFeedID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		MatchedBy:  matchedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create reconciliation match",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
			logger.Any("bank_feed_id", input.BankFeedID),
			logger.Any("target_type", input.TargetType),
			logger.Any("target_id", input.TargetID),
		)
		return 0, err
	}
	return res.LastInsertId()
}

// CountMatches returns the total number of reconciliation matches created for
// the owner, regardless of target type.
func (r *Repository) CountMatches(ctx context.Context, ownerID int64) (int64, error) {
	const q = `
		SELECT COUNT(*)
		FROM finance_reconciliation_matches
		WHERE owner_id = ?`

	var cnt int64
	if err := r.pool.QueryRowContext(ctx, q, ownerID).Scan(&cnt); err != nil {
		logger.Ctx(ctx).Error("Failed to count reconciliation matches", logger.Error(err))
		return 0, err
	}

	return cnt, nil
}

// CountMatchesByTargetType returns the number of matches for each target
// type ("invoice", "expense") for the given owner.
func (r *Repository) CountMatchesByTargetType(ctx context.Context, ownerID int64) (map[string]int64, error) {
	const q = `
		SELECT target_type, COUNT(*) AS cnt
		FROM finance_reconciliation_matches
		WHERE owner_id = ?
		GROUP BY target_type`

	rows, err := r.pool.QueryContext(ctx, q, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count matches by target type", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var ttype string
		var cnt int64
		if scanErr := rows.Scan(&ttype, &cnt); scanErr != nil {
			return nil, scanErr
		}
		counts[ttype] = cnt
	}

	return counts, rows.Err()
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

// decimalStrToMinorUnits converts a DECIMAL column (e.g. "12.34") into
// integer minor units (1234). Uses round-to-nearest (round-half-away-from-zero)
// via math.Round so the rounding mode is explicit and symmetric for positive
// and negative amounts. Falls back to 0 on parse failure (logged at Warn); the
// higher layer still gets the rest of the row so a single bad amount does not
// blackhole the whole list.
//
// NOTE: This uses round-to-nearest, NOT banker's rounding (round-to-even).
// All amounts in the system must use consistent rounding for reconciliation to work.
func decimalStrToMinorUnits(ctx context.Context, s sql.NullString) int64 {
	if !s.Valid || s.String == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to parse decimal amount from database",
			logger.String("value", s.String),
			logger.Error(err),
		)
		return 0
	}
	return int64(math.Round(f * 100))
}
