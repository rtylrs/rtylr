package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type UserRow struct {
	ID             int64
	Email          string
	FullName       sql.NullString
	BusinessID     sql.NullInt64
	IsConsoleAdmin bool
	IsBanned       bool
	BanReason      sql.NullString
	BannedAt       sql.NullTime
	CreatedAt      sql.NullTime
	PasswordHash   string
}

type BusinessRow struct {
	ID        int64
	Name      string
	OwnerID   int64
	CreatedAt sql.NullTime
}

type SubscriptionRow struct {
	ID                  int64
	OwnerID             int64
	PlanID              int64
	PlanName            string
	Status              string
	SeatCount           int
	TrialEndsAt         sql.NullTime
	CurrentPeriodEndsAt sql.NullTime
	CreatedAt           sql.NullTime
	OwnerEmail          sql.NullString
}

type PlanRow struct {
	ID                      int64
	Name                    string
	Slug                    string
	PricePerSeatCents       int
	AnnualPricePerSeatCents int
	TrialDays               int
	MaxSeats                sql.NullInt64
	MaxLocations            sql.NullInt64
	IsActive                bool
	IsSolo                  bool
	IsFeatured              bool
	TierOrder               int
	Tagline                 string
	CreatedAt               sql.NullTime
}

type GrantRow struct {
	ID          int64
	UserID      int64
	PlanID      int64
	PlanName    string
	GrantedBy   int64
	GranterName sql.NullString
	Note        sql.NullString
	ExpiresAt   sql.NullTime
	RevokedAt   sql.NullTime
	CreatedAt   sql.NullTime
	UserEmail   sql.NullString
}

type StatsRow struct {
	TotalUsers      int64
	BannedUsers     int64
	TotalBusinesses int64
	ActiveSubs      int64
	TrialSubs       int64
	ManualGrants    int64
}

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

var userColumns = []string{
	"id", "email", "full_name", "business_id", "is_console_admin",
	"is_banned", "ban_reason", "banned_at", "created_at", "password_hash",
}

// ── Auth ──────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*UserRow, error) {
	query, args, err := sq.
		Select(userColumns...).
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
		&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

// ── Users CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListUsers(ctx context.Context, limit, offset int, search string) ([]UserRow, int64, error) {
	conds := sq.And{}
	if search != "" {
		like := "%" + search + "%"
		conds = append(conds, sq.Or{
			sq.Like{"email": like},
			sq.Like{"full_name": like},
		})
	}

	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("users").
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
		Select(userColumns...).
		From("users").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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

	users := make([]UserRow, 0, limit)
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
			&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*UserRow, error) {
	query, args, err := sq.
		Select(userColumns...).
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
		&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string, isConsoleAdmin bool) (int64, error) {
	query, args, err := sq.
		Insert("users").
		Columns("email", "password_hash", "full_name", "is_console_admin").
		Values(email, passwordHash, fullName, isConsoleAdmin).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateUser(ctx context.Context, id int64, fullName, email string) error {
	query, args, err := sq.
		Update("users").
		Set("full_name", fullName).
		Set("email", email).
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

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) BanUser(ctx context.Context, id int64, reason string) error {
	query, args, err := sq.
		Update("users").
		Set("is_banned", true).
		Set("ban_reason", reason).
		Set("banned_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) UnbanUser(ctx context.Context, id int64) error {
	query, args, err := sq.
		Update("users").
		Set("is_banned", false).
		Set("ban_reason", nil).
		Set("banned_at", nil).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error {
	query, args, err := sq.
		Update("users").
		Set("is_console_admin", isAdmin).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Businesses CRUD ───────────────────────────────────────────────────────────

func (r *Repository) ListBusinesses(ctx context.Context, limit, offset int, search string) ([]BusinessRow, int64, error) {
	conds := sq.And{}
	if search != "" {
		conds = append(conds, sq.Like{"name": "%" + search + "%"})
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
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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
	query, args, err := sq.
		Select("id", "name", "owner_id", "created_at").
		From("businesses").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var b BusinessRow
	if err := row.Scan(&b.ID, &b.Name, &b.OwnerID, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) UpdateBusiness(ctx context.Context, id int64, name string) error {
	query, args, err := sq.
		Update("businesses").
		Set("name", name).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteBusiness(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("businesses").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Subscriptions CRUD ────────────────────────────────────────────────────────

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
		Select("s.id", "s.owner_id", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.current_period_ends_at", "s.created_at", "u.email").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		Where(conds).
		OrderBy("s.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
			&s.TrialEndsAt, &s.CurrentPeriodEndsAt, &s.CreatedAt, &s.OwnerEmail); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, total, rows.Err()
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id int64) (*SubscriptionRow, error) {
	query, args, err := sq.
		Select("s.id", "s.owner_id", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.current_period_ends_at", "s.created_at", "u.email").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		Where(sq.Eq{"s.id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var s SubscriptionRow
	if err := row.Scan(&s.ID, &s.OwnerID, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
		&s.TrialEndsAt, &s.CurrentPeriodEndsAt, &s.CreatedAt, &s.OwnerEmail); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionRow, error) {
	query, args, err := sq.
		Select("s.id", "s.owner_id", "s.plan_id", "p.name", "s.status", "s.seat_count",
			"s.trial_ends_at", "s.current_period_ends_at", "s.created_at", "u.email").
		From("subscriptions s").
		Join("plans p ON p.id = s.plan_id").
		LeftJoin("users u ON u.id = s.owner_id").
		Where(sq.Eq{"s.owner_id": ownerID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, query, args...)
	var s SubscriptionRow
	if err := row.Scan(&s.ID, &s.OwnerID, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
		&s.TrialEndsAt, &s.CurrentPeriodEndsAt, &s.CreatedAt, &s.OwnerEmail); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) UpsertSubscription(ctx context.Context, ownerID, planID int64, status string, seatCount int, expiresAt *time.Time) error {
	var expiresSQL interface{}
	if expiresAt != nil {
		expiresSQL = *expiresAt
	}
	query, args, err := sq.
		Insert("subscriptions").
		Columns("owner_id", "plan_id", "status", "seat_count", "current_period_ends_at").
		Values(ownerID, planID, status, seatCount, expiresSQL).
		Suffix(`ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), status = VALUES(status),
		   seat_count = VALUES(seat_count), current_period_ends_at = VALUES(current_period_ends_at),
		   updated_at = NOW()`).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) UpdateSubscription(ctx context.Context, id, planID int64, status string, seatCount int) error {
	query, args, err := sq.
		Update("subscriptions").
		Set("plan_id", planID).
		Set("status", status).
		Set("seat_count", seatCount).
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
	query, args, err := sq.
		Delete("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

// ── Plans CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListPlans(ctx context.Context) ([]PlanRow, error) {
	query, args, err := sq.
		Select("id", "name", "slug", "price_per_seat_cents", "annual_price_per_seat_cents", "trial_days",
			"max_seats", "max_locations", "is_active", "is_solo", "is_featured", "tier_order", "tagline", "created_at").
		From("plans").
		OrderBy("tier_order ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []PlanRow
	for rows.Next() {
		var p PlanRow
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.PricePerSeatCents, &p.AnnualPricePerSeatCents,
			&p.TrialDays, &p.MaxSeats, &p.MaxLocations, &p.IsActive, &p.IsSolo,
			&p.IsFeatured, &p.TierOrder, &p.Tagline, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
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
		Select("g.id", "g.user_id", "g.plan_id", "p.name", "g.granted_by", "adm.full_name",
			"g.note", "g.expires_at", "g.revoked_at", "g.created_at", "u.email").
		From("manual_access_grants g").
		Join("plans p ON p.id = g.plan_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("users adm ON adm.id = g.granted_by").
		OrderBy("g.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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
		if err := rows.Scan(&g.ID, &g.UserID, &g.PlanID, &g.PlanName, &g.GrantedBy, &g.GranterName,
			&g.Note, &g.ExpiresAt, &g.RevokedAt, &g.CreatedAt, &g.UserEmail); err != nil {
			return nil, 0, err
		}
		grants = append(grants, g)
	}
	return grants, total, rows.Err()
}

func (r *Repository) CreateGrant(ctx context.Context, userID, planID, grantedBy int64, note string, expiresAt *time.Time) (int64, error) {
	var expiresSQL interface{}
	if expiresAt != nil {
		expiresSQL = *expiresAt
	}
	query, args, err := sq.
		Insert("manual_access_grants").
		Columns("user_id", "plan_id", "granted_by", "note", "expires_at").
		Values(userID, planID, grantedBy, note, expiresSQL).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

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
	query, args, err := sq.
		Select(
			"(SELECT COUNT(*) FROM users) AS total_users",
			"(SELECT COUNT(*) FROM users WHERE is_banned = TRUE) AS banned_users",
			"(SELECT COUNT(*) FROM businesses) AS total_businesses",
			"(SELECT COUNT(*) FROM subscriptions WHERE status = 'active') AS active_subs",
			"(SELECT COUNT(*) FROM subscriptions WHERE status = 'trial') AS trial_subs",
			"(SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL) AS manual_grants",
		).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}

	var s StatsRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&s.TotalUsers, &s.BannedUsers, &s.TotalBusinesses, &s.ActiveSubs, &s.TrialSubs, &s.ManualGrants,
	); err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	return &s, nil
}
