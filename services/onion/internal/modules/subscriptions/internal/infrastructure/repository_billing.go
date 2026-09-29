package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) GetBillingHistory(ctx context.Context, ownerID int64, limit int) ([]domain.BillingHistoryMonth, error) {
	currentMonth := domain.FirstOfUTCMonth(time.Now().UTC())

	rows, err := r.queries.GetBillingHistoryPeriods(ctx, sqlc.GetBillingHistoryPeriodsParams{
		OwnerID:      ownerID,
		BillingMonth: currentMonth,
		Limit:        int32(limit * 20), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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

	daysInMonth := domain.DaysInUTCMonth(billingMonth)

	// A period closed before it started covers no day and bills nothing; see the
	// same rule in the application service layer.
	daysUsed := 0
	if !effectiveEnd.Before(start) {
		daysUsed = int(effectiveEnd.Sub(start).Hours()/24) + 1
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
