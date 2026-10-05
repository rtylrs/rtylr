package infrastructure

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionRow, error) {
	row, err := r.queries.GetSubscriptionByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	sub := subscriptionFromOwnerRow(row)
	if err := r.decorateSubscriptionSeats(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func nullableTimeValue(value NullTime) any {
	if value.Valid {
		return value.Time
	}
	return nil
}

func nullableStringValue(value NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func (r *Repository) UpsertSubscription(ctx context.Context, params SubscriptionWriteParams) error {
	return r.queries.UpsertSubscription(ctx, upsertSubscriptionParams(params))
}

// dynamic-query: squirrel
func (r *Repository) UpdateSubscription(ctx context.Context, id int64, params SubscriptionWriteParams) error {
	query, args, err := sq.
		Update("subscriptions").
		Set("plan_id", params.PlanID).
		Set("status", params.Status).
		Set("seat_count", params.SeatCount).
		Set("trial_ends_at", nullableTimeValue(params.TrialEndsAt)).
		Set("next_billing_date", nullableTimeValue(params.NextBillingDate)).
		Set("current_period_ends_at", nullableTimeValue(params.CurrentPeriodEndsAt)).
		Set("paddle_subscription_id", nullableStringValue(params.PaddleSubscriptionID)).
		Set("paddle_customer_id", nullableStringValue(params.PaddleCustomerID)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteSubscription(ctx context.Context, id int64) error {
	return r.queries.DeleteSubscription(ctx, id)
}

// ── Plans CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListPlans(ctx context.Context) ([]PlanRow, error) {
	rows, err := r.queries.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	plans := make([]PlanRow, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, planFromListRow(row))
	}
	return plans, nil
}

func (r *Repository) UpdatePlan(ctx context.Context, id int64, name string, pricePerSeatCents, annualPricePerSeatCents int, isActive bool) error {
	query, args, err := sq.
		Update("plans").
		Set("name", name).
		Set("price_per_seat_cents", pricePerSeatCents).
		Set("annual_price_per_seat_cents", annualPricePerSeatCents).
		Set("is_active", isActive).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Manual Access Grants CRUD ─────────────────────────────────────────────────

// dynamic-query: squirrel
func (r *Repository) ListGrants(ctx context.Context, limit, offset int) ([]GrantRow, int64, error) {
	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("manual_access_grants").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL, listArgs, err := sq.
		Select("g.id", "g.user_id", "COALESCE(u.full_name, '')", "g.plan_id", "p.name", "g.granted_by", "adm.full_name",
			"g.note", "g.expires_at", "g.revoked_at", "g.created_at", "u.email", "b.id", "b.name").
		From("manual_access_grants g").
		Join("plans p ON p.id = g.plan_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.owner_id = g.user_id AND COALESCE(b.is_active, TRUE) = TRUE").
		LeftJoin("users adm ON adm.id = g.granted_by").
		OrderBy("g.created_at DESC").
		Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	grants := make([]GrantRow, 0, limit)
	for rows.Next() {
		var g GrantRow
		if err := rows.Scan(&g.ID, &g.UserID, &g.UserName, &g.PlanID, &g.PlanName, &g.GrantedBy, &g.GranterName,
			&g.Note, &g.ExpiresAt, &g.RevokedAt, &g.CreatedAt, &g.UserEmail, &g.BusinessID, &g.BusinessName); err != nil {
			return nil, 0, err
		}
		if g.BusinessID.Valid {
			seatCount, err := r.CountActiveSeatsForOwner(ctx, g.UserID)
			if err != nil {
				return nil, 0, err
			}
			g.ActiveSeatCount = seatCount
		}
		grants = append(grants, g)
	}
	return grants, total, rows.Err()
}

func (r *Repository) CreateGrant(ctx context.Context, userID, planID, grantedBy int64, note string, expiresAt *time.Time) (int64, error) {
	query, args, err := buildCreateGrantSQL(userID, planID, grantedBy, note, expiresAt)
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func buildCreateGrantSQL(userID, planID, grantedBy int64, note string, expiresAt *time.Time) (string, []any, error) {
	var expiresSQL interface{}
	if expiresAt != nil {
		expiresSQL = *expiresAt
	}
	return sq.
		Insert("manual_access_grants").
		Columns("user_id", "plan_id", "granted_by", "note", "expires_at").
		Values(userID, planID, grantedBy, note, expiresSQL).
		PlaceholderFormat(sq.Question).
		ToSql()
}

// CreateGrantAndActivateSubscription inserts a manual access grant and mirrors
// it onto the owner's subscription row in one transaction. Entitlement is
// resolved from subscriptions, so a grant whose subscription write failed would
// report success while granting nothing; either both writes land or neither does.
func (r *Repository) CreateGrantAndActivateSubscription(ctx context.Context, userID, planID, grantedBy int64, note string, expiresAt *time.Time, sub SubscriptionWriteParams) (int64, error) {
	query, args, err := buildCreateGrantSQL(userID, planID, grantedBy, note, expiresAt)
	if err != nil {
		return 0, err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := sqlc.New(tx).UpsertSubscription(ctx, upsertSubscriptionParams(sub)); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// GetGrant returns the display fields for one grant without loading the full
// grant history. CreateGrant uses it to build its response efficiently.
func (r *Repository) GetGrant(ctx context.Context, id int64) (*GrantRow, error) {
	query, args, err := sq.
		Select("g.id", "g.user_id", "COALESCE(u.full_name, '')", "g.plan_id", "p.name", "g.granted_by", "adm.full_name",
			"g.note", "g.expires_at", "g.revoked_at", "g.created_at", "u.email", "b.id", "b.name").
		From("manual_access_grants g").
		Join("plans p ON p.id = g.plan_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.owner_id = g.user_id AND COALESCE(b.is_active, TRUE) = TRUE").
		LeftJoin("users adm ON adm.id = g.granted_by").
		Where(sq.Eq{"g.id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	var grant GrantRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&grant.ID, &grant.UserID, &grant.UserName, &grant.PlanID, &grant.PlanName, &grant.GrantedBy, &grant.GranterName,
		&grant.Note, &grant.ExpiresAt, &grant.RevokedAt, &grant.CreatedAt, &grant.UserEmail, &grant.BusinessID, &grant.BusinessName,
	); err != nil {
		return nil, err
	}
	if grant.BusinessID.Valid {
		seatCount, err := r.CountActiveSeatsForOwner(ctx, grant.UserID)
		if err != nil {
			return nil, err
		}
		grant.ActiveSeatCount = seatCount
	}
	return &grant, nil
}

func (r *Repository) GetGrantUserID(ctx context.Context, id int64) (int64, error) {
	query, args, err := sq.
		Select("user_id").
		From("manual_access_grants").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	var userID int64
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}

// dynamic-query: squirrel
func (r *Repository) RevokeGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Update("manual_access_grants").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("manual_access_grants").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Stats ─────────────────────────────────────────────────────────────────────

func (r *Repository) GetStats(ctx context.Context) (*StatsRow, error) {
	row, err := r.queries.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	return &StatsRow{
		TotalUsers:        row.TotalUsers,
		BannedUsers:       row.BannedUsers,
		TotalBusinesses:   row.TotalBusinesses,
		ActiveSubs:        row.ActiveSubs,
		TrialSubs:         row.TrialSubs,
		ManualGrants:      row.ManualGrants,
		ManualAddonGrants: row.ManualAddonGrants,
		ConsoleAdmins:     row.ConsoleAdmins,
	}, nil
}
