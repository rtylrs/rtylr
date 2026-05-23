package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

// planSelectCols is the canonical column list for reading a plan row.
// Kept in one place so the three GetPlan* variants stay in sync when columns change.
const planSelectCols = `id, name, slug, apps, price_per_seat_cents,
		COALESCE(annual_price_per_seat_cents, 0), trial_days,
		COALESCE(paddle_price_id, ''), COALESCE(tagline, ''),
		max_seats, max_locations,
		COALESCE(ai_credits_per_month, 0), COALESCE(features, JSON_ARRAY()),
		COALESCE(tier_order, 100), COALESCE(is_solo, FALSE), COALESCE(is_featured, FALSE),
		is_active, created_at`

// scanPlan hydrates a Plan from a row matching planSelectCols.
// Reused by all three GetPlan* entrypoints.
func scanPlan(scanner interface {
	Scan(dest ...any) error
},
) (domain.Plan, error) {
	var p domain.Plan
	var appsJSON, featuresJSON string
	var maxSeats, maxLocations sql.NullInt64
	err := scanner.Scan(
		&p.ID, &p.Name, &p.Slug, &appsJSON, &p.PricePerSeatCents,
		&p.AnnualPricePerSeatCents, &p.TrialDays,
		&p.PaddlePriceID, &p.Tagline,
		&maxSeats, &maxLocations,
		&p.AICreditsPerMonth, &featuresJSON,
		&p.TierOrder, &p.IsSolo, &p.IsFeatured,
		&p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		return domain.Plan{}, err
	}
	if jsonErr := json.Unmarshal([]byte(appsJSON), &p.Apps); jsonErr != nil {
		p.Apps = []string{}
	}
	if jsonErr := json.Unmarshal([]byte(featuresJSON), &p.Features); jsonErr != nil {
		p.Features = []string{}
	}
	if maxSeats.Valid {
		v := int(maxSeats.Int64)
		p.MaxSeats = &v
	}
	if maxLocations.Valid {
		v := int(maxLocations.Int64)
		p.MaxLocations = &v
	}
	return p, nil
}

func (r *Repository) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+planSelectCols+`
		 FROM plans WHERE is_active = TRUE
		 ORDER BY is_solo ASC, tier_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]domain.Plan, 0)
	for rows.Next() {
		p, scanErr := scanPlan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		plans = append(plans, p)
	}

	return plans, rows.Err()
}

func (r *Repository) GetPlanBySlug(ctx context.Context, slug string) (*domain.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+planSelectCols+` FROM plans WHERE slug = ? AND is_active = TRUE`,
		slug)
	p, err := scanPlan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, id int64) (*domain.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+planSelectCols+` FROM plans WHERE id = ?`,
		id)
	p, err := scanPlan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return &p, nil
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (r *Repository) GetSubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	var sub domain.Subscription
	var appsJSON string
	var trialEndsAt, periodEndsAt, nextBillingDate sql.NullTime
	var paddleSubID, paddleCustID sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.owner_id, s.plan_id, p.name, p.slug, p.apps, p.price_per_seat_cents,
		        s.status, s.seat_count, s.trial_ends_at, s.next_billing_date,
		        s.current_period_ends_at, s.paddle_subscription_id, s.paddle_customer_id, s.created_at
		 FROM subscriptions s
		 JOIN plans p ON p.id = s.plan_id
		 WHERE s.owner_id = ?`,
		ownerID).Scan(
		&sub.ID, &sub.OwnerID, &sub.PlanID, &sub.PlanName, &sub.PlanSlug, &appsJSON, &sub.PricePerSeatCents,
		&sub.Status, &sub.SeatCount, &trialEndsAt, &nextBillingDate,
		&periodEndsAt, &paddleSubID, &paddleCustID, &sub.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}

	if jsonErr := json.Unmarshal([]byte(appsJSON), &sub.Apps); jsonErr != nil {
		sub.Apps = []string{}
	}
	if trialEndsAt.Valid {
		t := trialEndsAt.Time
		sub.TrialEndsAt = &t
	}
	if nextBillingDate.Valid {
		t := nextBillingDate.Time
		sub.NextBillingDate = &t
	}
	if periodEndsAt.Valid {
		t := periodEndsAt.Time
		sub.CurrentPeriodEndsAt = &t
	}
	if paddleSubID.Valid {
		sub.PaddleSubscriptionID = paddleSubID.String
	}
	if paddleCustID.Valid {
		sub.PaddleCustomerID = paddleCustID.String
	}

	return &sub, nil
}

func (r *Repository) UpsertSubscription(ctx context.Context, ownerID, planID int64, trialDays int) error {
	trialEnd := time.Now().UTC().AddDate(0, 0, trialDays)
	// next_billing_date is always the 1st of next month
	now := time.Now().UTC()
	nextBilling := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (owner_id, plan_id, status, seat_count, trial_ends_at, next_billing_date)
		 VALUES (?, ?, 'trial', 1, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   plan_id = VALUES(plan_id),
		   updated_at = CURRENT_TIMESTAMP`,
		ownerID, planID, trialEnd, nextBilling)
	return err
}

func (r *Repository) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	var appsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT p.apps FROM subscriptions s JOIN plans p ON p.id = s.plan_id WHERE s.owner_id = ?`,
		ownerID).Scan(&appsJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}

	var apps []string
	if jsonErr := json.Unmarshal([]byte(appsJSON), &apps); jsonErr != nil {
		return []string{}, nil
	}
	return apps, nil
}

func (r *Repository) UpdateSeatCount(ctx context.Context, ownerID int64, seatCount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET seat_count = ?, updated_at = CURRENT_TIMESTAMP WHERE owner_id = ?`,
		seatCount, ownerID)
	return err
}

// ── Seat Periods ──────────────────────────────────────────────────────────────

func (r *Repository) GetActivePeriod(ctx context.Context, subscriptionID int64) (*domain.SeatPeriod, error) {
	var sp domain.SeatPeriod
	var periodEnd sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, subscription_id, owner_id, seat_count, price_per_seat_cents,
		        period_start, period_end, billing_month
		 FROM seat_periods
		 WHERE subscription_id = ? AND period_end IS NULL
		 ORDER BY period_start DESC LIMIT 1`,
		subscriptionID).Scan(
		&sp.ID, &sp.SubscriptionID, &sp.OwnerID, &sp.SeatCount, &sp.PricePerSeatCents,
		&sp.PeriodStart, &periodEnd, &sp.BillingMonth)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if periodEnd.Valid {
		t := periodEnd.Time
		sp.PeriodEnd = &t
	}
	return &sp, nil
}

func (r *Repository) CloseActivePeriod(ctx context.Context, subscriptionID int64, periodEnd time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE seat_periods SET period_end = ? WHERE subscription_id = ? AND period_end IS NULL`,
		periodEnd, subscriptionID)
	return err
}

func (r *Repository) OpenPeriod(ctx context.Context, subscriptionID, ownerID int64, seatCount, pricePerSeatCents int, periodStart, billingMonth time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO seat_periods (subscription_id, owner_id, seat_count, price_per_seat_cents, period_start, billing_month)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		subscriptionID, ownerID, seatCount, pricePerSeatCents, periodStart.Format("2006-01-02"), billingMonth.Format("2006-01-02"))
	return err
}

func (r *Repository) GetPeriodsForMonth(ctx context.Context, ownerID int64, month time.Time) ([]domain.SeatPeriod, error) {
	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, subscription_id, owner_id, seat_count, price_per_seat_cents,
		        period_start, period_end, billing_month
		 FROM seat_periods
		 WHERE owner_id = ? AND billing_month = ?
		 ORDER BY period_start ASC`,
		ownerID, billingMonth.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	periods := make([]domain.SeatPeriod, 0)
	for rows.Next() {
		var sp domain.SeatPeriod
		var periodEnd sql.NullTime
		if err := rows.Scan(&sp.ID, &sp.SubscriptionID, &sp.OwnerID, &sp.SeatCount, &sp.PricePerSeatCents,
			&sp.PeriodStart, &periodEnd, &sp.BillingMonth); err != nil {
			return nil, err
		}
		if periodEnd.Valid {
			t := periodEnd.Time
			sp.PeriodEnd = &t
		}
		periods = append(periods, sp)
	}
	return periods, rows.Err()
}

func (r *Repository) GetBillingHistory(ctx context.Context, ownerID int64, limit int) ([]domain.BillingHistoryMonth, error) {
	// Get distinct billing months that have closed periods (past months)
	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := r.db.QueryContext(ctx,
		`SELECT billing_month, seat_count, price_per_seat_cents, period_start, period_end
		 FROM seat_periods
		 WHERE owner_id = ? AND billing_month < ?
		 ORDER BY billing_month DESC, period_start ASC
		 LIMIT ?`,
		ownerID, currentMonth.Format("2006-01-02"), limit*20) // fetch enough rows for limit months
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monthMap := make(map[string]*domain.BillingHistoryMonth)
	monthOrder := make([]string, 0)

	for rows.Next() {
		var billingMonth time.Time
		var seatCount, pricePerSeatCents int
		var periodStart time.Time
		var periodEnd sql.NullTime

		if err := rows.Scan(&billingMonth, &seatCount, &pricePerSeatCents, &periodStart, &periodEnd); err != nil {
			return nil, err
		}

		key := billingMonth.Format("2006-01")
		if _, exists := monthMap[key]; !exists {
			m := &domain.BillingHistoryMonth{
				Month:  billingMonth,
				Status: "paid",
			}
			monthMap[key] = m
			monthOrder = append(monthOrder, key)
		}

		var endDate *time.Time
		if periodEnd.Valid {
			t := periodEnd.Time
			endDate = &t
		}

		item := computeLineItem(seatCount, pricePerSeatCents, periodStart, endDate, billingMonth)
		monthMap[key].LineItems = append(monthMap[key].LineItems, item)
		monthMap[key].TotalCents += item.AmountCents
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]domain.BillingHistoryMonth, 0, len(monthOrder))
	for i, key := range monthOrder {
		if i >= limit {
			break
		}
		result = append(result, *monthMap[key])
	}
	return result, nil
}

// computeLineItem calculates the amount for a seat period based on daily proration.
func computeLineItem(seatCount, pricePerSeatCents int, start time.Time, end *time.Time, billingMonth time.Time) domain.BillingLineItem {
	effectiveEnd := time.Now().UTC()
	if end != nil {
		effectiveEnd = *end
	}

	// days_in_month for this billing month
	firstOfNextMonth := time.Date(billingMonth.Year(), billingMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := int(firstOfNextMonth.Sub(billingMonth).Hours() / 24)

	// days used in this period
	daysUsed := int(effectiveEnd.Sub(start).Hours()/24) + 1
	if daysUsed < 1 {
		daysUsed = 1
	}
	if daysUsed > daysInMonth {
		daysUsed = daysInMonth
	}

	// amount = seat_count * days_used * (price_per_seat / days_in_month)
	// Use integer arithmetic: (seat_count * days_used * price_per_seat) / days_in_month
	amount := (seatCount * daysUsed * pricePerSeatCents) / daysInMonth

	return domain.BillingLineItem{
		PeriodStart:       start,
		PeriodEnd:         end,
		SeatCount:         seatCount,
		PricePerSeatCents: pricePerSeatCents,
		AmountCents:       amount,
	}
}

// ── Paddle ─────────────────────────────────────────────────────────────────────

// UpdatePaddleSubscription writes the Paddle subscription ID, customer ID, and status
// to the subscriptions row for the given owner.
func (r *Repository) UpdatePaddleSubscription(ctx context.Context, ownerID int64, paddleSubID, paddleCustID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions
		 SET paddle_subscription_id = ?, paddle_customer_id = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE owner_id = ?`,
		paddleSubID, paddleCustID, status, ownerID)
	return err
}

// GetOwnerIDByPaddleSubscription returns the owner_id for a given Paddle subscription ID.
func (r *Repository) GetOwnerIDByPaddleSubscription(ctx context.Context, paddleSubID string) (int64, error) {
	var ownerID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT owner_id FROM subscriptions WHERE paddle_subscription_id = ? LIMIT 1`,
		paddleSubID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}
