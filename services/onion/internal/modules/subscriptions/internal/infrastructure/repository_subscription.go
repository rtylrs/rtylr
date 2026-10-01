package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) UpsertSubscription(ctx context.Context, ownerID, planID int64, trialDays, seatCount int) error {
	status := sqlc.SubscriptionsStatusActive
	trialEnd := sql.NullTime{}
	if trialDays > 0 {
		status = sqlc.SubscriptionsStatusTrial
		trialEnd = sql.NullTime{Time: time.Now().UTC().AddDate(0, 0, trialDays), Valid: true}
	}
	if seatCount < 1 {
		seatCount = 1
	}
	now := time.Now().UTC()
	nextBilling := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	return r.queries.UpsertSubscription(ctx, sqlc.UpsertSubscriptionParams{
		OwnerID:         ownerID,
		PlanID:          planID,
		Status:          sqlc.NullSubscriptionsStatus{SubscriptionsStatus: status, Valid: true},
		SeatCount:       int32(seatCount),
		TrialEndsAt:     trialEnd,
		NextBillingDate: sql.NullTime{Time: nextBilling, Valid: true},
	})
}

func (r *Repository) CountActiveSeats(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountActiveSeats(ctx, sqlc.CountActiveSeatsParams{
		OwnerID:   ownerID,
		OwnerID_2: ownerID,
	})
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 1, nil
	}
	return int(count), nil
}

func (r *Repository) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	appsJSON, err := r.queries.GetOwnerApps(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}

	var apps []string
	if jsonErr := json.Unmarshal(appsJSON, &apps); jsonErr != nil {
		return []string{}, nil
	}
	return apps, nil
}

func (r *Repository) ExpireTrialSubscriptions(ctx context.Context) error {
	return r.queries.ExpireTrialSubscriptions(ctx)
}

// ── Seat Periods ──────────────────────────────────────────────────────────────

// ReconcileSeatBilling brings the seat ledger in line with the current seat
// count and the plan's current per-seat price, as of asOf. See
// domain.PlanSeatLedger for the policy; this method only applies it.
//
// This is the module's only writer of seat_periods and subscriptions.seat_count.
// The whole thing runs in one transaction that first ensures the subscription
// row exists and then locks it, so two concurrent callers serialise instead of
// both opening a period — the read path that used to do this work had no lock
// and could leave two open periods that are then both billed.
func (r *Repository) ReconcileSeatBilling(ctx context.Context, ownerID, defaultPlanID int64, defaultTrialDays, seatCount int, asOf time.Time) error {
	if seatCount < 1 {
		seatCount = 1
	}
	today := domain.UTCDay(asOf)
	nextBilling := domain.FirstOfUTCMonth(domain.FirstOfUTCMonth(today).AddDate(0, 1, 0))

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// ON DUPLICATE KEY UPDATE with a self-assignment only creates: an existing
	// subscription keeps its plan, status and seat count. It also takes the row
	// lock, so a concurrent creator waits instead of racing.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO subscriptions (owner_id, plan_id, status, seat_count, trial_ends_at, next_billing_date)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE owner_id = owner_id`,
		ownerID, defaultPlanID, newSubscriptionStatus(defaultTrialDays), seatCount,
		newTrialEnd(defaultTrialDays, asOf), nextBilling,
	); err != nil {
		return err
	}

	var subID int64
	var pricePerSeatCents int
	if err := tx.QueryRowContext(ctx, `
		SELECT s.id, p.price_per_seat_cents
		FROM subscriptions s
		JOIN plans p ON p.id = s.plan_id
		WHERE s.owner_id = ?
		FOR UPDATE`, ownerID).Scan(&subID, &pricePerSeatCents); err != nil {
		return err
	}

	open, err := loadOpenSeatPeriods(ctx, tx, subID)
	if err != nil {
		return err
	}

	plan := domain.PlanSeatLedger(open, today, seatCount, pricePerSeatCents)
	for _, c := range plan.CloseNoDay {
		if err := closeSeatPeriod(ctx, tx, c.ID, c.End); err != nil {
			return err
		}
	}
	if plan.CloseActive {
		if err := closeSeatPeriod(ctx, tx, plan.RunningID, plan.CloseRunning); err != nil {
			return err
		}
	}
	if plan.OpenNew {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO seat_periods (subscription_id, owner_id, seat_count, price_per_seat_cents, period_start, billing_month)
			VALUES (?, ?, ?, ?, ?, ?)`,
			subID, ownerID, seatCount, pricePerSeatCents, plan.NewStart, plan.NewBillingM,
		); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE subscriptions SET seat_count = ?, next_billing_date = ? WHERE id = ?`,
		seatCount, nextBilling, subID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func newSubscriptionStatus(trialDays int) string {
	if trialDays > 0 {
		return "trial"
	}
	return "active"
}

func newTrialEnd(trialDays int, asOf time.Time) sql.NullTime {
	if trialDays <= 0 {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: asOf.UTC().AddDate(0, 0, trialDays), Valid: true}
}

func loadOpenSeatPeriods(ctx context.Context, tx *sql.Tx, subscriptionID int64) ([]domain.SeatPeriod, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, seat_count, price_per_seat_cents, period_start, billing_month
		FROM seat_periods
		WHERE subscription_id = ? AND period_end IS NULL
		ORDER BY period_start ASC, id ASC`, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SeatPeriod
	for rows.Next() {
		var p domain.SeatPeriod
		p.SubscriptionID = subscriptionID
		if err := rows.Scan(&p.ID, &p.SeatCount, &p.PricePerSeatCents, &p.PeriodStart, &p.BillingMonth); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func closeSeatPeriod(ctx context.Context, tx *sql.Tx, periodID int64, periodEnd time.Time) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE seat_periods SET period_end = ? WHERE id = ?`, domain.UTCDay(periodEnd), periodID)
	return err
}

// ListOwnerIDsForReconciliation pages the owners whose ledger may be stale:
// every owner with an active business, an active permission grant, or an
// existing subscription row (an owner whose team was deactivated still has to
// be invoiced for the days they were active).
func (r *Repository) ListOwnerIDsForReconciliation(ctx context.Context, afterOwnerID int64, limit int) ([]int64, error) {
	if limit < 1 {
		limit = 1
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT owner_id FROM (
		    SELECT b.owner_id AS owner_id
		    FROM businesses b
		    WHERE COALESCE(b.is_active, TRUE) = TRUE
		    UNION
		    SELECT business.owner_id
		    FROM authorization_business_members member
		    JOIN businesses business ON business.id = member.business_id
		    WHERE member.status = 'active' AND COALESCE(business.is_active, TRUE) = TRUE
		    UNION
		    SELECT s.owner_id
		    FROM subscriptions s
		) candidates
		WHERE owner_id > ?
		ORDER BY owner_id ASC
		LIMIT ?`, afterOwnerID, limit) // #nosec G115 -- Reviewed width conversion is a validated application batch size.
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0, limit)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func seatPeriodFromRow(
	id, subscriptionID, ownerID int64,
	seatCount, pricePerSeatCents int,
	periodStart time.Time,
	periodEnd sql.NullTime,
	billingMonth time.Time,
) *domain.SeatPeriod {
	sp := &domain.SeatPeriod{
		ID:                id,
		SubscriptionID:    subscriptionID,
		OwnerID:           ownerID,
		SeatCount:         seatCount,
		PricePerSeatCents: pricePerSeatCents,
		PeriodStart:       periodStart,
		BillingMonth:      billingMonth,
	}
	if periodEnd.Valid {
		t := periodEnd.Time
		sp.PeriodEnd = &t
	}
	return sp
}

func (r *Repository) GetPeriodsForMonth(ctx context.Context, ownerID int64, month time.Time) ([]domain.SeatPeriod, error) {
	billingMonth := domain.FirstOfUTCMonth(month)
	rows, err := r.queries.GetSeatPeriodsForMonth(ctx, sqlc.GetSeatPeriodsForMonthParams{
		OwnerID:      ownerID,
		BillingMonth: billingMonth,
	})
	if err != nil {
		return nil, err
	}

	periods := make([]domain.SeatPeriod, 0, len(rows))
	for _, row := range rows {
		sp := seatPeriodFromRow(
			row.ID, row.SubscriptionID, row.OwnerID,
			int(row.SeatCount), int(row.PricePerSeatCents),
			row.PeriodStart, row.PeriodEnd, row.BillingMonth,
		)
		periods = append(periods, *sp)
	}
	return periods, nil
}

// dynamic-query: squirrel
func (r *Repository) GetAddonPeriodsForMonth(ctx context.Context, ownerID, businessID int64, month time.Time) ([]domain.AddonBillingPeriod, error) {
	if businessID <= 0 {
		return []domain.AddonBillingPeriod{}, nil
	}

	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := time.Date(billingMonth.Year(), billingMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	activeGrantPredicate := `
mag.user_id = ?
AND mag.business_id = ba.business_id
AND mag.addon_id = ba.addon_id
AND mag.revoked_at IS NULL
AND mag.created_at < ?
AND (mag.expires_at IS NULL OR mag.expires_at >= ?)`

	query, args, err := sq.
		Select("ba.addon_id", "a.name", "a.price", "ba.enabled_at", "ba.disabled_at").
		Column(sq.Expr("EXISTS (SELECT 1 FROM manual_addon_grants mag WHERE "+activeGrantPredicate+") AS is_granted", ownerID, firstOfNextMonth, billingMonth)).
		Column(sq.Expr("(SELECT mag.expires_at FROM manual_addon_grants mag WHERE "+activeGrantPredicate+" ORDER BY mag.created_at DESC LIMIT 1) AS grant_expires_at", ownerID, firstOfNextMonth, billingMonth)).
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
		var isGranted bool
		var grantExpiresAt sql.NullTime
		if err := rows.Scan(&addonID, &name, &price, &enabledAt, &disabledAt, &isGranted, &grantExpiresAt); err != nil {
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
		var grantExpires *time.Time
		if grantExpiresAt.Valid {
			t := grantExpiresAt.Time.UTC()
			grantExpires = &t
		}

		periods = append(periods, domain.AddonBillingPeriod{
			AddonID:        addonID,
			Name:           name,
			PriceCents:     decimalStringToCents(price),
			PeriodStart:    start,
			PeriodEnd:      end,
			BillingMonth:   billingMonth,
			IsGranted:      isGranted,
			GrantExpiresAt: grantExpires,
		})
	}

	return periods, rows.Err()
}
