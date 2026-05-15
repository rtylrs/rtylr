package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/database/dbclient"
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

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

// ── Auth ──────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*UserRow, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, email, full_name, business_id, is_console_admin, is_banned, ban_reason, banned_at, created_at, password_hash
		 FROM users WHERE email = ? LIMIT 1`, email)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
		&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

// ── Users CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListUsers(ctx context.Context, limit, offset int, search string) ([]UserRow, int64, error) {
	base := `FROM users WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		base += ` AND (email LIKE ? OR full_name LIKE ?)`
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, email, full_name, business_id, is_console_admin, is_banned, ban_reason, banned_at, created_at, password_hash ` +
		base + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []UserRow
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
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, email, full_name, business_id, is_console_admin, is_banned, ban_reason, banned_at, created_at, password_hash
		 FROM users WHERE id = ? LIMIT 1`, id)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.BusinessID,
		&u.IsConsoleAdmin, &u.IsBanned, &u.BanReason, &u.BannedAt, &u.CreatedAt, &u.PasswordHash); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string, isConsoleAdmin bool) (int64, error) {
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, full_name, is_console_admin) VALUES (?, ?, ?, ?)`,
		email, passwordHash, fullName, isConsoleAdmin)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateUser(ctx context.Context, id int64, fullName, email string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE users SET full_name = ?, email = ?, updated_at = NOW() WHERE id = ?`,
		fullName, email, id)
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (r *Repository) BanUser(ctx context.Context, id int64, reason string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE users SET is_banned = TRUE, ban_reason = ?, banned_at = NOW() WHERE id = ?`,
		reason, id)
	return err
}

func (r *Repository) UnbanUser(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE users SET is_banned = FALSE, ban_reason = NULL, banned_at = NULL WHERE id = ?`, id)
	return err
}

func (r *Repository) SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE users SET is_console_admin = ? WHERE id = ?`, isAdmin, id)
	return err
}

// ── Businesses CRUD ───────────────────────────────────────────────────────────

func (r *Repository) ListBusinesses(ctx context.Context, limit, offset int, search string) ([]BusinessRow, int64, error) {
	base := `FROM businesses WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		base += ` AND name LIKE ?`
		args = append(args, "%"+search+"%")
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, owner_id, created_at ` + base + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var businesses []BusinessRow
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
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, name, owner_id, created_at FROM businesses WHERE id = ? LIMIT 1`, id)
	var b BusinessRow
	if err := row.Scan(&b.ID, &b.Name, &b.OwnerID, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) UpdateBusiness(ctx context.Context, id int64, name string) error {
	_, err := r.pool.ExecContext(ctx, `UPDATE businesses SET name = ? WHERE id = ?`, name, id)
	return err
}

func (r *Repository) DeleteBusiness(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx, `DELETE FROM businesses WHERE id = ?`, id)
	return err
}

// ── Subscriptions CRUD ────────────────────────────────────────────────────────

func (r *Repository) ListSubscriptions(ctx context.Context, limit, offset int, status string) ([]SubscriptionRow, int64, error) {
	base := `FROM subscriptions s JOIN plans p ON p.id = s.plan_id LEFT JOIN users u ON u.id = s.owner_id WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		base += ` AND s.status = ?`
		args = append(args, status)
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT s.id, s.owner_id, s.plan_id, p.name, s.status, s.seat_count,
		s.trial_ends_at, s.current_period_ends_at, s.created_at, u.email ` +
		base + ` ORDER BY s.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subs []SubscriptionRow
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
	row := r.pool.QueryRowContext(ctx,
		`SELECT s.id, s.owner_id, s.plan_id, p.name, s.status, s.seat_count,
		 s.trial_ends_at, s.current_period_ends_at, s.created_at, u.email
		 FROM subscriptions s JOIN plans p ON p.id = s.plan_id LEFT JOIN users u ON u.id = s.owner_id
		 WHERE s.id = ? LIMIT 1`, id)
	var s SubscriptionRow
	if err := row.Scan(&s.ID, &s.OwnerID, &s.PlanID, &s.PlanName, &s.Status, &s.SeatCount,
		&s.TrialEndsAt, &s.CurrentPeriodEndsAt, &s.CreatedAt, &s.OwnerEmail); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) GetSubscriptionByOwner(ctx context.Context, ownerID int64) (*SubscriptionRow, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT s.id, s.owner_id, s.plan_id, p.name, s.status, s.seat_count,
		 s.trial_ends_at, s.current_period_ends_at, s.created_at, u.email
		 FROM subscriptions s JOIN plans p ON p.id = s.plan_id LEFT JOIN users u ON u.id = s.owner_id
		 WHERE s.owner_id = ? LIMIT 1`, ownerID)
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
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO subscriptions (owner_id, plan_id, status, seat_count, current_period_ends_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), status = VALUES(status),
		   seat_count = VALUES(seat_count), current_period_ends_at = VALUES(current_period_ends_at),
		   updated_at = NOW()`,
		ownerID, planID, status, seatCount, expiresSQL)
	return err
}

func (r *Repository) UpdateSubscription(ctx context.Context, id, planID int64, status string, seatCount int) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE subscriptions SET plan_id = ?, status = ?, seat_count = ?, updated_at = NOW() WHERE id = ?`,
		planID, status, seatCount, id)
	return err
}

func (r *Repository) DeleteSubscription(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = ?`, id)
	return err
}

// ── Plans CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) ListPlans(ctx context.Context) ([]PlanRow, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,
		 max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at
		 FROM plans ORDER BY tier_order ASC`)
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
	_, err := r.pool.ExecContext(ctx,
		`UPDATE plans SET name = ?, price_per_seat_cents = ?, annual_price_per_seat_cents = ?, is_active = ? WHERE id = ?`,
		name, pricePerSeatCents, annualPricePerSeatCents, isActive, id)
	return err
}

// ── Manual Access Grants CRUD ─────────────────────────────────────────────────

func (r *Repository) ListGrants(ctx context.Context, limit, offset int) ([]GrantRow, int64, error) {
	var total int64
	if err := r.pool.QueryRowContext(ctx, `SELECT COUNT(*) FROM manual_access_grants`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx,
		`SELECT g.id, g.user_id, g.plan_id, p.name, g.granted_by, adm.full_name,
		 g.note, g.expires_at, g.revoked_at, g.created_at, u.email
		 FROM manual_access_grants g
		 JOIN plans p ON p.id = g.plan_id
		 LEFT JOIN users u ON u.id = g.user_id
		 LEFT JOIN users adm ON adm.id = g.granted_by
		 ORDER BY g.created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var grants []GrantRow
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
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO manual_access_grants (user_id, plan_id, granted_by, note, expires_at) VALUES (?, ?, ?, ?, ?)`,
		userID, planID, grantedBy, note, expiresSQL)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) RevokeGrant(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE manual_access_grants SET revoked_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *Repository) DeleteGrant(ctx context.Context, id int64) error {
	_, err := r.pool.ExecContext(ctx, `DELETE FROM manual_access_grants WHERE id = ?`, id)
	return err
}

// ── Stats ─────────────────────────────────────────────────────────────────────

func (r *Repository) GetStats(ctx context.Context) (*StatsRow, error) {
	var s StatsRow
	err := r.pool.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM users WHERE is_banned = TRUE) as banned_users,
			(SELECT COUNT(*) FROM businesses) as total_businesses,
			(SELECT COUNT(*) FROM subscriptions WHERE status = 'active') as active_subs,
			(SELECT COUNT(*) FROM subscriptions WHERE status = 'trial') as trial_subs,
			(SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL) as manual_grants
	`).Scan(&s.TotalUsers, &s.BannedUsers, &s.TotalBusinesses, &s.ActiveSubs, &s.TrialSubs, &s.ManualGrants)
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	return &s, nil
}
