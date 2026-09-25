package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/billing/internal/domain"
)

var ErrBillingInvoicesUnavailable = errors.New("billing invoice history is not available")

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

// GetSummary reads the same subscription row that gates the rest of the
// product. A tenant without a row is still on the seeded free plan until the
// subscription service creates its row; paid state must never be replaced by
// a hard-coded response.
func (r *Repository) GetSummary(ctx context.Context, ownerID int64) (domain.BillingSummary, error) {
	if ownerID <= 0 {
		return domain.BillingSummary{}, fmt.Errorf("invalid billing owner id %d", ownerID)
	}

	var (
		planSlug string
		status   string
		seats    int
		maxSeats sql.NullInt64
		nextDate sql.NullTime
	)
	err := r.pool.QueryRowContext(ctx, `
		SELECT p.slug, s.status, s.seat_count, p.max_seats, s.next_billing_date
		FROM subscriptions s
		JOIN plans p ON p.id = s.plan_id
		WHERE s.owner_id = ?`, ownerID).Scan(&planSlug, &status, &seats, &maxSeats, &nextDate)
	if errors.Is(err, sql.ErrNoRows) {
		return r.freeSummary(ctx)
	}
	if err != nil {
		return domain.BillingSummary{}, err
	}

	summary := domain.BillingSummary{
		Plan:         planSlug,
		Status:       status,
		CurrentUsage: float64(max(seats, 0)),
		UsageLimit:   nullableLimit(maxSeats),
	}
	if nextDate.Valid {
		date := nextDate.Time
		summary.NextBillingDate = &date
	}
	return summary, nil
}

func (r *Repository) freeSummary(ctx context.Context) (domain.BillingSummary, error) {
	var maxSeats sql.NullInt64
	err := r.pool.QueryRowContext(ctx, `
		SELECT max_seats
		FROM plans
		WHERE slug = 'free' AND is_active = TRUE
		LIMIT 1`).Scan(&maxSeats)
	if err != nil {
		return domain.BillingSummary{}, err
	}
	return domain.BillingSummary{
		Plan:       "free",
		Status:     "active",
		UsageLimit: nullableLimit(maxSeats),
	}, nil
}

func nullableLimit(value sql.NullInt64) float64 {
	if !value.Valid || value.Int64 < 0 {
		return 0
	}
	return float64(value.Int64)
}

// ListInvoices has no authoritative billing-invoice table in this deployment.
// Returning an empty slice would turn a missing provider integration into a
// successful statement for every tenant, so fail explicitly until an invoice
// source is wired in.
func (r *Repository) ListInvoices(_ context.Context, ownerID int64) ([]domain.Invoice, error) {
	if ownerID <= 0 {
		return nil, fmt.Errorf("invalid billing owner id %d", ownerID)
	}
	return nil, ErrBillingInvoicesUnavailable
}
