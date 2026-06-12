package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

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
var planSelectCols = []string{
	"id", "name", "slug", "apps", "price_per_seat_cents",
	"COALESCE(annual_price_per_seat_cents, 0)", "trial_days",
	"COALESCE(paddle_price_id, '')", "COALESCE(tagline, '')",
	"max_seats", "max_locations",
	"COALESCE(ai_credits_per_month, 0)", "COALESCE(features, JSON_ARRAY())",
	"COALESCE(tier_order, 100)", "COALESCE(is_solo, FALSE)", "COALESCE(is_featured, FALSE)",
	"is_active", "created_at",
}

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
	query, args, err := sq.
		Select(planSelectCols...).
		From("plans").
		Where(sq.Eq{"is_active": true}).
		OrderBy("is_solo ASC", "tier_order ASC", "id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
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
	query, args, err := sq.
		Select(planSelectCols...).
		From("plans").
		Where(sq.Eq{"slug": slug, "is_active": true}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)
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
	query, args, err := sq.
		Select(planSelectCols...).
		From("plans").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)
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

	query, args, err := sq.
		Select(
			"s.id", "s.owner_id", "s.plan_id", "p.name", "p.slug", "p.apps", "p.price_per_seat_cents",
			"s.status", "s.seat_count", "s.trial_ends_at", "s.next_billing_date",
			"s.current_period_ends_at", "s.paddle_subscription_id", "s.paddle_customer_id", "s.created_at",
		).
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		Where(sq.Eq{"s.owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
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

	query, args, err := sq.
		Insert("subscriptions").
		Columns("owner_id", "plan_id", "status", "seat_count", "trial_ends_at", "next_billing_date").
		Values(ownerID, planID, "trial", 1, trialEnd, nextBilling).
		Suffix("ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), updated_at = CURRENT_TIMESTAMP").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	var appsJSON string
	query, args, err := sq.
		Select("p.apps").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		Where(sq.Eq{"s.owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&appsJSON)
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
	query, args, err := sq.
		Update("subscriptions").
		Set("seat_count", seatCount).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// ── Seat Periods ──────────────────────────────────────────────────────────────

func (r *Repository) GetActivePeriod(ctx context.Context, subscriptionID int64) (*domain.SeatPeriod, error) {
	var sp domain.SeatPeriod
	var periodEnd sql.NullTime

	query, args, err := sq.
		Select(
			"id", "subscription_id", "owner_id", "seat_count", "price_per_seat_cents",
			"period_start", "period_end", "billing_month",
		).
		From("seat_periods").
		Where(sq.Eq{"subscription_id": subscriptionID, "period_end": nil}).
		OrderBy("period_start DESC").
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
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
	query, args, err := sq.
		Update("seat_periods").
		Set("period_end", periodEnd).
		Where(sq.Eq{"subscription_id": subscriptionID, "period_end": nil}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) OpenPeriod(ctx context.Context, subscriptionID, ownerID int64, seatCount, pricePerSeatCents int, periodStart, billingMonth time.Time) error {
	query, args, err := sq.
		Insert("seat_periods").
		Columns("subscription_id", "owner_id", "seat_count", "price_per_seat_cents", "period_start", "billing_month").
		Values(subscriptionID, ownerID, seatCount, pricePerSeatCents, periodStart.Format("2006-01-02"), billingMonth.Format("2006-01-02")).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetPeriodsForMonth(ctx context.Context, ownerID int64, month time.Time) ([]domain.SeatPeriod, error) {
	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	query, args, err := sq.
		Select(
			"id", "subscription_id", "owner_id", "seat_count", "price_per_seat_cents",
			"period_start", "period_end", "billing_month",
		).
		From("seat_periods").
		Where(sq.Eq{"owner_id": ownerID, "billing_month": billingMonth.Format("2006-01-02")}).
		OrderBy("period_start ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *Repository) GetAddonPeriodsForMonth(ctx context.Context, businessID int64, month time.Time) ([]domain.AddonBillingPeriod, error) {
	if businessID <= 0 {
		return []domain.AddonBillingPeriod{}, nil
	}

	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := time.Date(billingMonth.Year(), billingMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	query, args, err := sq.
		Select("ba.addon_id", "a.name", "a.price", "ba.enabled_at", "ba.disabled_at").
		From("business_addons ba").
		Join("addons a ON a.id = ba.addon_id").
		Where(sq.Eq{"ba.business_id": businessID, "a.is_free": false}).
		Where(sq.Gt{"a.price": 0}).
		Where(sq.Lt{"ba.enabled_at": firstOfNextMonth}).
		Where(sq.Or{
			sq.Eq{"ba.disabled_at": nil},
			sq.GtOrEq{"ba.disabled_at": billingMonth},
		}).
		OrderBy("ba.enabled_at ASC", "ba.addon_id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	periods := make([]domain.AddonBillingPeriod, 0)
	for rows.Next() {
		var addonID, name, price string
		var enabledAt time.Time
		var disabledAt sql.NullTime
		if err := rows.Scan(&addonID, &name, &price, &enabledAt, &disabledAt); err != nil {
			return nil, err
		}

		start := enabledAt.UTC()
		if start.Before(billingMonth) {
			start = billingMonth
		}

		var end *time.Time
		if disabledAt.Valid {
			disabled := disabledAt.Time.UTC()
			if disabled.Before(firstOfNextMonth) {
				end = &disabled
			}
		}

		periods = append(periods, domain.AddonBillingPeriod{
			AddonID:      addonID,
			Name:         name,
			PriceCents:   decimalStringToCents(price),
			PeriodStart:  start,
			PeriodEnd:    end,
			BillingMonth: billingMonth,
		})
	}

	return periods, rows.Err()
}

func (r *Repository) GetBillingHistory(ctx context.Context, ownerID int64, limit int) ([]domain.BillingHistoryMonth, error) {
	// Get distinct billing months that have closed periods (past months)
	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// fetch enough rows for limit months
	query, args, err := sq.
		Select("billing_month", "seat_count", "price_per_seat_cents", "period_start", "period_end").
		From("seat_periods").
		Where(sq.Eq{"owner_id": ownerID}).
		Where(sq.Lt{"billing_month": currentMonth.Format("2006-01-02")}).
		OrderBy("billing_month DESC", "period_start ASC").
		Limit(uint64(limit * 20)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monthMap := make(map[string]*domain.BillingHistoryMonth, limit)
	monthOrder := make([]string, 0, limit)

	for rows.Next() {
		var billingMonth time.Time
		var seatCount, pricePerSeatCents int
		var periodStart time.Time
		var periodEnd sql.NullTime

		if err := rows.Scan(&billingMonth, &seatCount, &pricePerSeatCents, &periodStart, &periodEnd); err != nil {
			return nil, err
		}

		key := billingMonth.Format("2006-01")
		m, exists := monthMap[key]
		if !exists {
			m = &domain.BillingHistoryMonth{
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
		m.LineItems = append(m.LineItems, item)
		m.TotalCents += item.AmountCents
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
		Type:              "seat",
		Description:       "Seat usage",
		Quantity:          seatCount,
		UnitPriceCents:    pricePerSeatCents,
	}
}

func decimalStringToCents(value string) int {
	if value == "" {
		return 0
	}
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return int(math.Round(amount * 100))
}

// ── Paddle ─────────────────────────────────────────────────────────────────────

// UpdatePaddleSubscription writes the Paddle subscription ID, customer ID, and status
// to the subscriptions row for the given owner.
func (r *Repository) UpdatePaddleSubscription(ctx context.Context, ownerID int64, paddleSubID, paddleCustID, status string) error {
	query, args, err := sq.
		Update("subscriptions").
		Set("paddle_subscription_id", paddleSubID).
		Set("paddle_customer_id", paddleCustID).
		Set("status", status).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// GetOwnerIDByPaddleSubscription returns the owner_id for a given Paddle subscription ID.
func (r *Repository) GetOwnerIDByPaddleSubscription(ctx context.Context, paddleSubID string) (int64, error) {
	var ownerID int64
	query, args, err := sq.
		Select("owner_id").
		From("subscriptions").
		Where(sq.Eq{"paddle_subscription_id": paddleSubID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}
