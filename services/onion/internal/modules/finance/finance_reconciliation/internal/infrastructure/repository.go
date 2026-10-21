package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListUnmatchedBankFeeds(ctx context.Context, businessID int64) ([]domain.UnmatchedBankTransaction, error) {
	rows, err := r.queries.ListUnmatchedBankFeeds(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list unmatched bank feeds",
			logger.Error(err),
			logger.Any("business_id", businessID),
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
func (r *Repository) ListUnmatchedInvoices(ctx context.Context, businessID int64) ([]domain.UnmatchedInvoice, error) {
	// cross-schema read from finance_invoices — reconciliation needs the outstanding invoice side of the match (owned by the invoices module).
	rows, err := r.queries.ListUnmatchedInvoices(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list unmatched invoices",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)
		return nil, err
	}

	items := make([]domain.UnmatchedInvoice, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.UnmatchedInvoice{
			ID:               row.ID,
			InvoiceNumber:    row.InvoiceNumber,
			ClientName:       row.ClientName,
			AmountMinorUnits: decimalStrToMinorUnits(ctx, row.Total),
			Currency:         nullStr(row.Currency),
			IssueDate:        row.IssueDate,
			DueDate:          row.DueDate,
		})
	}
	return items, nil
}

func (r *Repository) PopulateOutstandingInvoiceAmounts(ctx context.Context, businessID int64, invoices []domain.UnmatchedInvoice) error {
	if len(invoices) == 0 {
		return nil
	}

	placeholders := make([]string, len(invoices))
	args := make([]any, 0, len(invoices)+1)
	args = append(args, businessID)
	for i, invoice := range invoices {
		placeholders[i] = "?"
		args = append(args, invoice.ID)
	}

	query := `
		SELECT id, CAST(ROUND(COALESCE(amount_due, 0) * 100) AS SIGNED)
		FROM finance_invoices
		WHERE business_id = ? AND id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	outstanding := make(map[int64]int64, len(invoices))
	for rows.Next() {
		var id, amountMinorUnits int64
		if err := rows.Scan(&id, &amountMinorUnits); err != nil {
			return err
		}
		outstanding[id] = amountMinorUnits
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range invoices {
		if amountMinorUnits, ok := outstanding[invoices[i].ID]; ok {
			invoices[i].AmountMinorUnits = amountMinorUnits
		}
	}
	return nil
}

// ListUnmatchedExpenses reads from finance_expenses. Same cross-schema
// rationale as ListUnmatchedInvoices. The expenses module owns the table but
// reconciliation legitimately needs to see the payable side for match
// suggestions.
func (r *Repository) ListUnmatchedExpenses(ctx context.Context, businessID int64) ([]domain.UnmatchedInvoice, error) {
	// cross-schema read from finance_expenses — reconciliation needs the outstanding expense side of the match (owned by the expenses module).
	rows, err := r.queries.ListUnmatchedExpenses(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list unmatched expenses",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)
		return nil, err
	}

	items := make([]domain.UnmatchedInvoice, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.UnmatchedInvoice{
			ID:               row.ID,
			InvoiceNumber:    row.Description,
			ClientName:       row.Vendor,
			AmountMinorUnits: decimalStrToMinorUnits(ctx, sql.NullString{String: row.Amount, Valid: true}),
			Currency:         nullStr(row.Currency),
			IssueDate:        row.ExpenseDate,
			DueDate:          row.ExpenseDate,
		})
	}
	return items, nil
}

func (r *Repository) BankFeedExists(ctx context.Context, businessID, id int64) (bool, error) {
	_, err := r.queries.GetBankFeedForOwner(ctx, sqlc.GetBankFeedForOwnerParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) TargetExists(ctx context.Context, businessID int64, input domain.MatchInput) (bool, error) {
	return targetExists(ctx, r.queries, businessID, input)
}

func targetExists(ctx context.Context, queries *sqlc.Queries, businessID int64, input domain.MatchInput) (bool, error) {
	var err error
	switch input.TargetType {
	case domain.TargetTypeInvoice:
		_, err = queries.GetInvoiceForBusiness(ctx, sqlc.GetInvoiceForBusinessParams{
			ID: input.TargetID, BusinessID: businessID,
		})
	case domain.TargetTypeExpense:
		_, err = queries.GetExpenseForBusiness(ctx, sqlc.GetExpenseForBusinessParams{
			ID: input.TargetID, BusinessID: businessID,
		})
	default:
		return false, domain.ErrInvalidTarget
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) CreateMatch(ctx context.Context, businessID int64, input domain.MatchInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	queries := r.queries.WithTx(tx)

	if _, err := queries.LockBankFeedForBusiness(ctx, sqlc.LockBankFeedForBusinessParams{
		ID: input.BankFeedID, BusinessID: businessID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrBankFeedNotFound
		}
		return 0, err
	}

	if _, err := queries.GetMatchForBankFeed(ctx, sqlc.GetMatchForBankFeedParams{
		BusinessID: businessID, BankFeedID: input.BankFeedID,
	}); err == nil {
		return 0, domain.ErrAlreadyMatched
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if err := lockTarget(ctx, queries, businessID, input); err != nil {
		return 0, err
	}
	if _, err := queries.GetMatchForTarget(ctx, sqlc.GetMatchForTargetParams{
		BusinessID: businessID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
	}); err == nil {
		return 0, domain.ErrAlreadyMatched
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	reservation, err := queries.ReserveReconciliationTarget(ctx, sqlc.ReserveReconciliationTargetParams{
		BusinessID: businessID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
	})
	if err != nil {
		return 0, err
	}
	reserved, err := reservation.RowsAffected()
	if err != nil {
		return 0, err
	}
	if reserved == 0 {
		return 0, domain.ErrAlreadyMatched
	}

	matchedBy := sql.NullInt64{}
	if input.MatchedBy > 0 {
		matchedBy = sql.NullInt64{Int64: input.MatchedBy, Valid: true}
	}
	res, err := queries.CreateReconciliationMatch(ctx, sqlc.CreateReconciliationMatchParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		BankFeedID: input.BankFeedID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		MatchedBy:  matchedBy,
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return 0, domain.ErrAlreadyMatched
		}
		logger.Ctx(ctx).Error("Failed to create reconciliation match",
			logger.Error(err),
			logger.Any("business_id", businessID),
			logger.Any("bank_feed_id", input.BankFeedID),
			logger.Any("target_type", input.TargetType),
			logger.Any("target_id", input.TargetID),
		)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Keep the denormalized bank-feed projection in lockstep with the
	// reconciliation ledger. This update is deliberately inside the same
	// transaction as CreateReconciliationMatch: a failure here rolls back the
	// match rather than leaving a successful-looking row that still appears
	// unmatched to every bank-feed consumer.
	var projection sql.Result
	switch input.TargetType {
	case domain.TargetTypeInvoice:
		projection, err = queries.SetBankFeedInvoiceMatch(ctx, sqlc.SetBankFeedInvoiceMatchParams{
			MatchedInvoiceID: sql.NullInt64{Int64: input.TargetID, Valid: true},
			ID:               input.BankFeedID,
			BusinessID:       businessID,
		})
	case domain.TargetTypeExpense:
		projection, err = queries.SetBankFeedExpenseMatch(ctx, sqlc.SetBankFeedExpenseMatchParams{
			MatchedExpenseID: sql.NullInt64{Int64: input.TargetID, Valid: true},
			ID:               input.BankFeedID,
			BusinessID:       businessID,
		})
	default:
		return 0, domain.ErrInvalidTarget
	}
	if err != nil {
		return 0, err
	}
	updated, err := projection.RowsAffected()
	if err != nil {
		return 0, err
	}
	if updated != 1 {
		return 0, domain.ErrAlreadyMatched
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func lockTarget(ctx context.Context, queries *sqlc.Queries, businessID int64, input domain.MatchInput) error {
	var err error
	switch input.TargetType {
	case domain.TargetTypeInvoice:
		_, err = queries.LockInvoiceForBusiness(ctx, sqlc.LockInvoiceForBusinessParams{
			ID: input.TargetID, BusinessID: businessID,
		})
	case domain.TargetTypeExpense:
		_, err = queries.LockExpenseForBusiness(ctx, sqlc.LockExpenseForBusinessParams{
			ID: input.TargetID, BusinessID: businessID,
		})
	default:
		return domain.ErrInvalidTarget
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrTargetNotFound
	}
	return err
}

// CountMatches returns the total number of reconciliation matches created for
// the owner, regardless of target type.
func (r *Repository) CountMatches(ctx context.Context, businessID int64) (int64, error) {
	cnt, err := r.queries.CountMatches(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count reconciliation matches", logger.Error(err))
		return 0, err
	}

	return cnt, nil
}

// CountMatchesByTargetType returns the number of matches for each target
// type ("invoice", "expense") for the given owner.
func (r *Repository) CountMatchesByTargetType(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountMatchesByTargetType(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count matches by target type", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		counts[row.TargetType] = row.Cnt
	}

	return counts, nil
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
