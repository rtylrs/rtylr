package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

func hydratePlanJSON(p *domain.Plan, appsJSON, featuresJSON string, maxSeats, maxLocations sql.NullInt64) {
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
}

func planFromSlugRow(row sqlc.GetPlanBySlugRow) domain.Plan {
	return planFromFields(
		row.ID, row.Name, row.Slug, row.Apps, row.PricePerSeatCents, row.AnnualPricePerSeatCents,
		row.TrialDays, row.PaddlePriceID, row.Tagline, row.MaxSeats, row.MaxLocations,
		row.AiCreditsPerMonth, row.Features, row.TierOrder, row.IsSolo, row.IsFeatured,
		row.IsActive, row.CreatedAt,
	)
}

func planFromIDRow(row sqlc.GetPlanByIDRow) domain.Plan {
	return planFromFields(
		row.ID, row.Name, row.Slug, row.Apps, row.PricePerSeatCents, row.AnnualPricePerSeatCents,
		row.TrialDays, row.PaddlePriceID, row.Tagline, row.MaxSeats, row.MaxLocations,
		row.AiCreditsPerMonth, row.Features, row.TierOrder, row.IsSolo, row.IsFeatured,
		row.IsActive, row.CreatedAt,
	)
}

func planFromListRow(row sqlc.ListPlansRow) domain.Plan {
	return planFromFields(
		row.ID, row.Name, row.Slug, row.Apps, row.PricePerSeatCents, row.AnnualPricePerSeatCents,
		row.TrialDays, row.PaddlePriceID, row.Tagline, row.MaxSeats, row.MaxLocations,
		row.AiCreditsPerMonth, row.Features, row.TierOrder, row.IsSolo, row.IsFeatured,
		row.IsActive, row.CreatedAt,
	)
}

func planFromFields(
	id int64, name, slug string, apps json.RawMessage,
	pricePerSeatCents, annualPricePerSeatCents, trialDays int32,
	paddlePriceID, tagline string,
	maxSeats, maxLocations sql.NullInt32,
	aiCreditsPerMonth int32, features json.RawMessage,
	tierOrder int32, isSolo, isFeatured bool,
	isActive sql.NullBool, createdAt sql.NullTime,
) domain.Plan {
	var p domain.Plan
	p.ID = id
	p.Name = name
	p.Slug = slug
	p.PricePerSeatCents = int(pricePerSeatCents)
	p.AnnualPricePerSeatCents = int(annualPricePerSeatCents)
	p.TrialDays = int(trialDays)
	p.PaddlePriceID = paddlePriceID
	p.Tagline = tagline
	p.AICreditsPerMonth = int(aiCreditsPerMonth)
	p.TierOrder = int(tierOrder)
	p.IsSolo = isSolo
	p.IsFeatured = isFeatured
	if isActive.Valid {
		p.IsActive = isActive.Bool
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	var maxSeatsNull, maxLocationsNull sql.NullInt64
	if maxSeats.Valid {
		maxSeatsNull = sql.NullInt64{Int64: int64(maxSeats.Int32), Valid: true}
	}
	if maxLocations.Valid {
		maxLocationsNull = sql.NullInt64{Int64: int64(maxLocations.Int32), Valid: true}
	}
	hydratePlanJSON(&p, string(apps), string(features), maxSeatsNull, maxLocationsNull)
	return p
}

func (r *Repository) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	rows, err := r.queries.ListPlans(ctx)
	if err != nil {
		return nil, err
	}

	plans := make([]domain.Plan, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, planFromListRow(row))
	}
	return plans, nil
}

func (r *Repository) GetPlanBySlug(ctx context.Context, slug string) (*domain.Plan, error) {
	row, err := r.queries.GetPlanBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	p := planFromSlugRow(row)
	return &p, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, id int64) (*domain.Plan, error) {
	row, err := r.queries.GetPlanByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	p := planFromIDRow(row)
	return &p, nil
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (r *Repository) GetSubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	row, err := r.queries.GetSubscription(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}

	sub := &domain.Subscription{
		ID:                row.ID,
		OwnerID:           row.OwnerID,
		PlanID:            row.PlanID,
		PlanName:          row.PlanName,
		PlanSlug:          row.PlanSlug,
		PricePerSeatCents: int(row.PricePerSeatCents),
		SeatCount:         int(row.SeatCount),
	}
	if row.Status.Valid {
		sub.Status = string(row.Status.SubscriptionsStatus)
	}
	if jsonErr := json.Unmarshal(row.Apps, &sub.Apps); jsonErr != nil {
		sub.Apps = []string{}
	}
	if row.TrialEndsAt.Valid {
		t := row.TrialEndsAt.Time
		sub.TrialEndsAt = &t
	}
	if row.NextBillingDate.Valid {
		t := row.NextBillingDate.Time
		sub.NextBillingDate = &t
	}
	if row.CurrentPeriodEndsAt.Valid {
		t := row.CurrentPeriodEndsAt.Time
		sub.CurrentPeriodEndsAt = &t
	}
	if row.PaddleSubscriptionID.Valid {
		sub.PaddleSubscriptionID = row.PaddleSubscriptionID.String
	}
	if row.PaddleCustomerID.Valid {
		sub.PaddleCustomerID = row.PaddleCustomerID.String
	}
	if row.CreatedAt.Valid {
		sub.CreatedAt = row.CreatedAt.Time
	}

	if err := r.hydrateActiveGrant(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *Repository) hydrateActiveGrant(ctx context.Context, sub *domain.Subscription) error {
	expiresAt, err := r.queries.GetActiveManualGrantExpiresAt(ctx, sqlc.GetActiveManualGrantExpiresAtParams{
		UserID: sub.OwnerID,
		PlanID: sub.PlanID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	sub.IsGranted = true
	if expiresAt.Valid {
		t := expiresAt.Time
		sub.GrantExpiresAt = &t
	}
	return nil
}

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
		SeatCount: int32(seatCount),
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
		SeatCount:         int32(seatCount),
		PricePerSeatCents: int32(pricePerSeatCents),
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

func (r *Repository) GetBillingHistory(ctx context.Context, ownerID int64, limit int) ([]domain.BillingHistoryMonth, error) {
	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := r.queries.GetBillingHistoryPeriods(ctx, sqlc.GetBillingHistoryPeriodsParams{
		OwnerID:      ownerID,
		BillingMonth: currentMonth,
		Limit:        int32(limit * 20),
	})
	if err != nil {
		return nil, err
	}

	monthMap := make(map[string]*domain.BillingHistoryMonth, limit)
	monthOrder := make([]string, 0, limit)

	for _, row := range rows {
		key := row.BillingMonth.Format("2006-01")
		m, exists := monthMap[key]
		if !exists {
			m = &domain.BillingHistoryMonth{
				Month:  row.BillingMonth,
				Status: "paid",
			}
			monthMap[key] = m
			monthOrder = append(monthOrder, key)
		}

		var endDate *time.Time
		if row.PeriodEnd.Valid {
			t := row.PeriodEnd.Time
			endDate = &t
		}

		item := computeLineItem(
			int(row.SeatCount),
			int(row.PricePerSeatCents),
			row.PeriodStart,
			endDate,
			row.BillingMonth,
		)
		m.LineItems = append(m.LineItems, item)
		m.TotalCents += item.AmountCents
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

	firstOfNextMonth := time.Date(billingMonth.Year(), billingMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := int(firstOfNextMonth.Sub(billingMonth).Hours() / 24)

	daysUsed := int(effectiveEnd.Sub(start).Hours()/24) + 1
	if daysUsed < 1 {
		daysUsed = 1
	}
	if daysUsed > daysInMonth {
		daysUsed = daysInMonth
	}

	includedSeats, billableSeats := splitSeats(seatCount, domain.FreeIncludedSeatLimit)
	amount := (billableSeats * daysUsed * pricePerSeatCents) / daysInMonth
	includedAmount := (includedSeats * daysUsed * pricePerSeatCents) / daysInMonth

	return domain.BillingLineItem{
		PeriodStart:       start,
		PeriodEnd:         end,
		SeatCount:         seatCount,
		IncludedSeatLimit: domain.FreeIncludedSeatLimit,
		IncludedSeatCount: includedSeats,
		BillableSeatCount: billableSeats,
		IncludedCents:     includedAmount,
		PricePerSeatCents: pricePerSeatCents,
		AmountCents:       amount,
		Type:              "seat",
		Description:       "Seat usage",
		Quantity:          seatCount,
		UnitPriceCents:    pricePerSeatCents,
	}
}

func splitSeats(seatCount, includedLimit int) (int, int) {
	if seatCount < 0 {
		seatCount = 0
	}
	if includedLimit < 0 {
		includedLimit = 0
	}
	included := seatCount
	if included > includedLimit {
		included = includedLimit
	}
	return included, seatCount - included
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

// UpdatePaddleSubscription reports whether a row matched: the first webhook
// for an account can arrive before the lazily-created subscription row does,
// and a silently-dropped update leaves a paying customer on free access (#755).
func (r *Repository) UpdatePaddleSubscription(ctx context.Context, ownerID int64, paddleSubID, paddleCustID, status string) (bool, error) {
	res, err := r.queries.UpdatePaddleSubscription(ctx, sqlc.UpdatePaddleSubscriptionParams{
		PaddleSubscriptionID: sql.NullString{String: paddleSubID, Valid: paddleSubID != ""},
		PaddleCustomerID:     sql.NullString{String: paddleCustID, Valid: paddleCustID != ""},
		Status: sqlc.NullSubscriptionsStatus{
			SubscriptionsStatus: sqlc.SubscriptionsStatus(status),
			Valid:               status != "",
		},
		OwnerID: ownerID,
	})
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ApplyPaddleSubscriptionEvent atomically deduplicates and orders a webhook
// with the subscription transition. A failed update rolls back the ledger
// insert so Paddle can retry when a subscription row is created later.
func (r *Repository) ApplyPaddleSubscriptionEvent(ctx context.Context, eventID, paddleSubID string, occurredAt time.Time, fallbackOwnerID int64, paddleCustID, status string) (bool, int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return false, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var lastOccurred sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT occurred_at
		FROM paddle_webhook_events
		WHERE paddle_subscription_id = ?
		ORDER BY occurred_at DESC
		LIMIT 1
		FOR UPDATE`, paddleSubID).Scan(&lastOccurred)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, 0, err
	}
	if err == nil && lastOccurred.Valid && occurredAt.Before(lastOccurred.Time) {
		if _, err := tx.ExecContext(ctx, `
			INSERT IGNORE INTO paddle_webhook_events (event_id, paddle_subscription_id, occurred_at)
			VALUES (?, ?, ?)`, eventID, paddleSubID, occurredAt); err != nil {
			return false, 0, err
		}
		if err := tx.Commit(); err != nil {
			return false, 0, err
		}
		return false, 0, nil
	}

	qtx := r.queries.WithTx(tx)
	res, err := qtx.InsertPaddleEventIfNew(ctx, sqlc.InsertPaddleEventIfNewParams{
		EventID:              eventID,
		PaddleSubscriptionID: paddleSubID,
		OccurredAt:           occurredAt,
	})
	if err != nil {
		return false, 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, 0, err
	}
	if n == 0 {
		if err := tx.Commit(); err != nil {
			return false, 0, err
		}
		return false, 0, nil
	}

	ownerID, ownerErr := qtx.GetOwnerIDByPaddleSubscription(ctx, sql.NullString{String: paddleSubID, Valid: true})
	if errors.Is(ownerErr, sql.ErrNoRows) {
		ownerID = fallbackOwnerID
		ownerErr = nil
	}
	if ownerErr != nil {
		return false, 0, ownerErr
	}
	if ownerID == 0 {
		return false, 0, fmt.Errorf("cannot resolve owner from paddle subscription %s", paddleSubID)
	}

	updateResult, err := qtx.UpdatePaddleSubscription(ctx, sqlc.UpdatePaddleSubscriptionParams{
		PaddleSubscriptionID: sql.NullString{String: paddleSubID, Valid: true},
		PaddleCustomerID:     sql.NullString{String: paddleCustID, Valid: paddleCustID != ""},
		Status: sqlc.NullSubscriptionsStatus{
			SubscriptionsStatus: sqlc.SubscriptionsStatus(status),
			Valid:               status != "",
		},
		OwnerID: ownerID,
	})
	if err != nil {
		return false, 0, err
	}
	updated, err := updateResult.RowsAffected()
	if err != nil {
		return false, 0, err
	}
	if updated == 0 {
		return false, 0, fmt.Errorf("no subscription row for owner %d yet; retry after local creation", ownerID)
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return true, ownerID, nil
}

// RecordPaddleEvent inserts the event id into the dedupe ledger and reports
// whether it was new. Duplicate deliveries return false so handlers can ack
// without re-applying a transition.
func (r *Repository) RecordPaddleEvent(ctx context.Context, eventID, paddleSubID string, occurredAt time.Time) (bool, error) {
	res, err := r.queries.InsertPaddleEventIfNew(ctx, sqlc.InsertPaddleEventIfNewParams{
		EventID:              eventID,
		PaddleSubscriptionID: paddleSubID,
		OccurredAt:           occurredAt,
	})
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// LastPaddleEventAt returns the newest event time already applied for the
// subscription, for out-of-order delivery rejection. The MAX() aggregate is
// scanned as interface{} by sqlc; the driver yields time.Time under
// parseTime DSNs, with string/[]byte fallbacks for other configurations.
func (r *Repository) LastPaddleEventAt(ctx context.Context, paddleSubID string) (time.Time, bool, error) {
	raw, err := r.queries.MaxPaddleEventOccurredAt(ctx, paddleSubID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	switch v := raw.(type) {
	case nil:
		return time.Time{}, false, nil
	case time.Time:
		return v, true, nil
	case []byte:
		parsed, perr := time.Parse("2006-01-02 15:04:05", string(v))
		if perr != nil {
			return time.Time{}, false, perr
		}
		return parsed, true, nil
	case string:
		parsed, perr := time.Parse("2006-01-02 15:04:05", v)
		if perr != nil {
			return time.Time{}, false, perr
		}
		return parsed, true, nil
	default:
		return time.Time{}, false, fmt.Errorf("unexpected paddle event timestamp type %T", raw)
	}
}

func (r *Repository) GetOwnerIDByPaddleSubscription(ctx context.Context, paddleSubID string) (int64, error) {
	return r.queries.GetOwnerIDByPaddleSubscription(ctx, sql.NullString{String: paddleSubID, Valid: paddleSubID != ""})
}
