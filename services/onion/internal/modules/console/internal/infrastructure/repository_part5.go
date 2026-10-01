package infrastructure

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) BanUser(ctx context.Context, id int64, reason string) error {
	return r.queries.BanUser(ctx, sqlc.BanUserParams{
		BanReason: sql.NullString{String: reason, Valid: true},
		ID:        id,
	})
}

func (r *Repository) UnbanUser(ctx context.Context, id int64) error {
	return r.queries.UnbanUser(ctx, id)
}

func (r *Repository) SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error {
	return r.queries.SetConsoleAdmin(ctx, sqlc.SetConsoleAdminParams{
		IsConsoleAdmin: isAdmin,
		ID:             id,
	})
}

// ── Businesses CRUD ───────────────────────────────────────────────────────────

// dynamic-query: squirrel
func (r *Repository) ListBusinesses(
	ctx context.Context,
	limit, offset int,
	id int64,
	name string,
	ownerID int64,
) ([]BusinessRow, int64, error) {
	conds := sq.And{}
	if id > 0 {
		conds = append(conds, sq.Eq{"id": id})
	}
	if name != "" {
		conds = append(conds, sq.Eq{"name": name})
	}
	if ownerID > 0 {
		conds = append(conds, sq.Eq{"owner_id": ownerID})
	}

	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("businesses").
		Where(conds).
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
		Select("id", "name", "owner_id", "created_at").
		From("businesses").
		Where(conds).
		OrderBy("created_at DESC").
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

	businesses := make([]BusinessRow, 0, limit)
	for rows.Next() {
		var b BusinessRow
		if err := rows.Scan(&b.ID, &b.Name, &b.OwnerID, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		businesses = append(businesses, b)
	}
	return businesses, total, rows.Err()
}

func (r *Repository) GetBusinessByID(ctx context.Context, id int64) (*BusinessRow, error) {
	row, err := r.queries.GetBusinessByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return businessFromRow(row.ID, row.Name, row.OwnerID, row.CreatedAt), nil
}

func (r *Repository) GetActiveBusinessByOwner(ctx context.Context, ownerID int64) (*BusinessRow, error) {
	row, err := r.queries.GetActiveBusinessByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return businessFromRow(row.ID, row.Name, row.OwnerID, row.CreatedAt), nil
}

func (r *Repository) UpdateBusiness(ctx context.Context, id int64, name string) error {
	return r.queries.UpdateBusiness(ctx, sqlc.UpdateBusinessParams{
		Name: name,
		ID:   id,
	})
}

func (r *Repository) DeleteBusiness(ctx context.Context, id int64) error {
	return r.queries.DeleteBusiness(ctx, id)
}

func (r *Repository) CountActiveSeatsForOwner(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountActiveSeatsForOwner(ctx, sqlc.CountActiveSeatsForOwnerParams{
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

func (r *Repository) decorateSubscriptionSeats(ctx context.Context, sub *SubscriptionRow) error {
	activeSeats, err := r.CountActiveSeatsForOwner(ctx, sub.OwnerID)
	if err != nil {
		return err
	}
	sub.ActiveSeatCount = activeSeats
	sub.IncludedSeatCount = activeSeats
	if sub.IncludedSeatCount > 10 {
		sub.IncludedSeatCount = 10
	}
	sub.BillableSeatCount = activeSeats - sub.IncludedSeatCount
	return nil
}

// ── Subscriptions CRUD ────────────────────────────────────────────────────────

// dynamic-query: squirrel
func (r *Repository) ListSubscriptions(ctx context.Context, limit, offset int, status string) ([]SubscriptionRow, int64, error) {
	conds := sq.And{}
	if status != "" {
		conds = append(conds, sq.Eq{"s.status": status})
	}

	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		Where(conds).
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
		Select("s.id", "s.owner_id", "COALESCE(u.full_name, '')", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.next_billing_date", "s.current_period_ends_at", "s.paddle_subscription_id", "s.paddle_customer_id",
			"s.created_at", "s.updated_at", "u.email", "b.id", "b.name").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		LeftJoin("businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE").
		Where(conds).
		OrderBy("s.created_at DESC").
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

	subs := make([]SubscriptionRow, 0, limit)
	for rows.Next() {
		var s SubscriptionRow
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.OwnerName, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
			&s.TrialEndsAt, &s.NextBillingDate, &s.CurrentPeriodEndsAt, &s.PaddleSubscriptionID, &s.PaddleCustomerID,
			&s.CreatedAt, &s.UpdatedAt, &s.OwnerEmail, &s.BusinessID, &s.BusinessName); err != nil {
			return nil, 0, err
		}
		if err := r.decorateSubscriptionSeats(ctx, &s); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, total, rows.Err()
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id int64) (*SubscriptionRow, error) {
	row, err := r.queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	sub := subscriptionFromIDRow(row)
	if err := r.decorateSubscriptionSeats(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}
