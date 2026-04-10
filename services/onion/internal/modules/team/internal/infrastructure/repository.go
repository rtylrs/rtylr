package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// ─── Permissions ─────────────────────────────────────────────────────────────

func (r *Repository) GetPermission(ctx context.Context, ownerID, userID int64) (*domain.UserAppPermission, error) {
	query := `
		SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
		FROM user_app_permissions
		WHERE owner_id = ? AND user_id = ?`

	row := r.pool.QueryRowContext(ctx, query, ownerID, userID)

	return scanPermission(row)
}

func (r *Repository) GetPermissionByUserID(ctx context.Context, userID int64) (*domain.UserAppPermission, error) {
	query := `
		SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
		FROM user_app_permissions
		WHERE user_id = ?
		LIMIT 1`

	row := r.pool.QueryRowContext(ctx, query, userID)

	perm, err := scanPermission(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return perm, nil
}

func (r *Repository) ListPermissions(ctx context.Context, ownerID int64) ([]domain.UserAppPermission, error) {
	query := `
		SELECT id, owner_id, user_id, apps, role, is_active, created_at, updated_at
		FROM user_app_permissions
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list user app permissions", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var perms []domain.UserAppPermission

	for rows.Next() {
		p, scanErr := scanPermissionRow(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan permission row", logger.Error(scanErr))

			return nil, scanErr
		}

		perms = append(perms, p)
	}

	return perms, rows.Err()
}

func (r *Repository) UpsertPermission(ctx context.Context, ownerID, userID int64, apps []string, role domain.Role) error {
	appsJSON, err := json.Marshal(apps)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user_app_permissions (owner_id, user_id, apps, role)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE apps = VALUES(apps), role = VALUES(role), updated_at = NOW()`

	_, err = r.pool.ExecContext(ctx, query, ownerID, userID, appsJSON, string(role))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert user app permission", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeactivateUser(ctx context.Context, ownerID, userID int64) error {
	query := `
		UPDATE user_app_permissions
		SET is_active = FALSE, updated_at = NOW()
		WHERE owner_id = ? AND user_id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to deactivate user", logger.Error(err))

		return err
	}

	return nil
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (int64, string, string, error) {
	var id int64
	var storedEmail, fullName sql.NullString

	err := r.pool.QueryRowContext(ctx,
		`SELECT id, email, COALESCE(full_name,'') FROM users WHERE email = ? LIMIT 1`,
		email,
	).Scan(&id, &storedEmail, &fullName)
	if err != nil {
		return 0, "", "", err
	}

	return id, storedEmail.String, fullName.String, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (string, string, error) {
	var email string
	var fullName sql.NullString

	err := r.pool.QueryRowContext(ctx,
		`SELECT email, COALESCE(full_name,'') FROM users WHERE id = ?`,
		id,
	).Scan(&email, &fullName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", domain.ErrUserNotFound
		}

		return "", "", err
	}

	return email, fullName.String, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string) (int64, error) {
	res, err := r.pool.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?)`,
		email, passwordHash, sql.NullString{String: fullName, Valid: fullName != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user via hub", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

// ─── Staff listing (join users + permissions) ─────────────────────────────────

func (r *Repository) ListStaffUsers(ctx context.Context, ownerID int64) ([]domain.StaffUser, error) {
	query := `
		SELECT u.id, u.email, COALESCE(u.full_name,''), u.created_at,
		       p.apps, p.role, p.is_active
		FROM user_app_permissions p
		JOIN users u ON u.id = p.user_id
		WHERE p.owner_id = ?
		ORDER BY u.created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff users", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var users []domain.StaffUser

	for rows.Next() {
		var su domain.StaffUser
		var appsJSON []byte
		var role string

		if err := rows.Scan(&su.ID, &su.Email, &su.FullName, &su.CreatedAt, &appsJSON, &role, &su.IsActive); err != nil {
			logger.Ctx(ctx).Error("Failed to scan staff user row", logger.Error(err))

			return nil, err
		}

		su.Role = domain.Role(role)

		if err := json.Unmarshal(appsJSON, &su.Apps); err != nil {
			su.Apps = []string{}
		}

		users = append(users, su)
	}

	return users, rows.Err()
}

// ─── helpers ─────────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPermission(row rowScanner) (*domain.UserAppPermission, error) {
	var p domain.UserAppPermission
	var appsJSON []byte
	var role string

	if err := row.Scan(&p.ID, &p.OwnerID, &p.UserID, &appsJSON, &role, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}

	p.Role = domain.Role(role)

	if err := json.Unmarshal(appsJSON, &p.Apps); err != nil {
		p.Apps = []string{}
	}

	return &p, nil
}

func scanPermissionRow(rows *sql.Rows) (domain.UserAppPermission, error) {
	var p domain.UserAppPermission
	var appsJSON []byte
	var role string

	if err := rows.Scan(&p.ID, &p.OwnerID, &p.UserID, &appsJSON, &role, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.UserAppPermission{}, err
	}

	p.Role = domain.Role(role)

	if err := json.Unmarshal(appsJSON, &p.Apps); err != nil {
		p.Apps = []string{}
	}

	return p, nil
}

// ─── Invite Tokens ──────────────────────────────────────────────────────────

func (r *Repository) CreateInviteToken(ctx context.Context, userID, ownerID int64, token string, expiresAt time.Time) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO invite_tokens (user_id, owner_id, token, expires_at) VALUES (?, ?, ?, ?)`,
		userID, ownerID, token, expiresAt,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invite token", logger.Error(err))
	}
	return err
}

func (r *Repository) GetInviteToken(ctx context.Context, token string) (int64, int64, time.Time, error) {
	var userID, ownerID int64
	var expiresAt time.Time
	var usedAt sql.NullTime

	err := r.pool.QueryRowContext(ctx,
		`SELECT user_id, owner_id, expires_at, used_at FROM invite_tokens WHERE token = ?`,
		token,
	).Scan(&userID, &ownerID, &expiresAt, &usedAt)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	if usedAt.Valid {
		return 0, 0, time.Time{}, domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(expiresAt) {
		return 0, 0, time.Time{}, domain.ErrTokenExpired
	}

	return userID, ownerID, expiresAt, nil
}

func (r *Repository) MarkInviteTokenUsed(ctx context.Context, token string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE invite_tokens SET used_at = NOW() WHERE token = ?`,
		token,
	)
	return err
}

func (r *Repository) SetPassword(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID,
	)
	return err
}

// GetBusinessByID returns the business name and owner ID.
func (r *Repository) GetBusinessByID(ctx context.Context, businessID int64) (string, int64, error) {
	var name string
	var ownerID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT name, owner_id FROM businesses WHERE id = ?`,
		businessID,
	).Scan(&name, &ownerID)
	if err != nil {
		return "", 0, err
	}
	return name, ownerID, nil
}
