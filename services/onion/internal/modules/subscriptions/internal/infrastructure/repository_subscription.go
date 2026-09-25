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
		OwnerID_3: ownerID,
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

func (r *Repository) UpdateSeatCount(ctx context.Context, ownerID int64, seatCount int) error {
	return r.queries.UpdateSeatCount(ctx, sqlc.UpdateSeatCountParams{
		SeatCount: int32(seatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		OwnerID:   ownerID,
	})
}

// ── Seat Periods ──────────────────────────────────────────────────────────────

func (r *Repository) GetActivePeriod(ctx context.Context, subscriptionID int64) (*domain.SeatPeriod, error) {
	row, err := r.queries.GetActiveSeatPeriod(ctx, subscriptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return seatPeriodFromRow(
		row.ID, row.SubscriptionID, row.OwnerID,
		int(row.SeatCount), int(row.PricePerSeatCents),
		row.PeriodStart, row.PeriodEnd, row.BillingMonth,
	), nil
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

func (r *Repository) CloseActivePeriod(ctx context.Context, subscriptionID int64, periodEnd time.Time) error {
	return r.queries.CloseActiveSeatPeriod(ctx, sqlc.CloseActiveSeatPeriodParams{
		PeriodEnd:      sql.NullTime{Time: periodEnd, Valid: true},
		SubscriptionID: subscriptionID,
	})
}

func (r *Repository) OpenPeriod(ctx context.Context, subscriptionID, ownerID int64, seatCount, pricePerSeatCents int, periodStart, billingMonth time.Time) error {
	return r.queries.OpenSeatPeriod(ctx, sqlc.OpenSeatPeriodParams{
		SubscriptionID:    subscriptionID,
		OwnerID:           ownerID,
		SeatCount:         int32(seatCount),         // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PricePerSeatCents: int32(pricePerSeatCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PeriodStart:       periodStart,
		BillingMonth:      billingMonth,
	})
}

func (r *Repository) GetPeriodsForMonth(ctx context.Context, ownerID int64, month time.Time) ([]domain.SeatPeriod, error) {
	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
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
