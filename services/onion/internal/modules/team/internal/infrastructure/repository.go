package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

// ─── Permissions ─────────────────────────────────────────────────────────────

func (r *Repository) GetPermission(ctx context.Context, ownerID, userID int64) (*domain.UserAppPermission, error) {
	row, err := r.queries.GetPermission(ctx, sqlc.GetPermissionParams{
		OwnerID: ownerID,
		UserID:  userID,
	})
	if err != nil {
		return nil, err
	}

	return permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) GetPermissionByBusiness(ctx context.Context, businessID, userID int64) (*domain.UserAppPermission, error) {
	row, err := r.queries.GetPermissionByBusiness(ctx, sqlc.GetPermissionByBusinessParams{
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	return permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) GetPermissionByUserID(ctx context.Context, userID int64) (*domain.UserAppPermission, error) {
	row, err := r.queries.GetPermissionByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) ListPermissions(ctx context.Context, ownerID int64) ([]domain.UserAppPermission, error) {
	rows, err := r.queries.ListPermissions(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list user app permissions", logger.Error(err))

		return nil, err
	}

	var perms []domain.UserAppPermission

	for _, row := range rows {
		p := permissionFromRow(row.ID, row.OwnerID, row.UserID, row.Apps, row.Role, row.IsActive, row.CreatedAt, row.UpdatedAt)
		perms = append(perms, *p)
	}

	return perms, nil
}

func (r *Repository) UpsertPermission(ctx context.Context, ownerID, userID int64, apps []string, role domain.Role) error {
	appsJSON, err := json.Marshal(apps)
	if err != nil {
		return err
	}

	err = r.queries.UpsertPermission(ctx, sqlc.UpsertPermissionParams{
		OwnerID: ownerID,
		UserID:  userID,
		Apps:    appsJSON,
		Role:    sqlc.NullUserAppPermissionsRole{UserAppPermissionsRole: sqlc.UserAppPermissionsRole(role), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert user app permission", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeactivateUser(ctx context.Context, ownerID, userID int64) error {
	err := r.queries.DeactivateUser(ctx, sqlc.DeactivateUserParams{
		OwnerID: ownerID,
		UserID:  userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to deactivate user", logger.Error(err))

		return err
	}

	return nil
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (int64, string, string, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return 0, "", "", err
	}

	return row.ID, row.Email, row.FullName, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (string, string, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", domain.ErrUserNotFound
		}

		return "", "", err
	}

	return row.Email, row.FullName, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string) (int64, error) {
	res, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     sql.NullString{String: fullName, Valid: fullName != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create user via hub", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

// ─── Staff listing (join users + permissions) ─────────────────────────────────

func (r *Repository) ListStaffUsers(ctx context.Context, ownerID int64) ([]domain.StaffUser, error) {
	rows, err := r.queries.ListStaffUsers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff users", logger.Error(err))

		return nil, err
	}

	var users []domain.StaffUser

	for _, row := range rows {
		var su domain.StaffUser
		su.ID = row.ID
		su.Email = row.Email
		su.FullName = row.FullName
		if row.CreatedAt.Valid {
			su.CreatedAt = row.CreatedAt.Time
		}
		su.Role = domain.Role(row.Role.UserAppPermissionsRole)
		if row.IsActive.Valid {
			su.IsActive = row.IsActive.Bool
		}

		if err := json.Unmarshal(row.Apps, &su.Apps); err != nil {
			su.Apps = []string{}
		}

		users = append(users, su)
	}

	return users, nil
}

// ─── Invite Tokens ──────────────────────────────────────────────────────────

func (r *Repository) CreateInviteToken(ctx context.Context, userID, ownerID int64, token string, expiresAt time.Time) error {
	err := r.queries.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{
		UserID:    userID,
		OwnerID:   ownerID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invite token", logger.Error(err))
	}
	return err
}

func (r *Repository) GetInviteToken(ctx context.Context, token string) (int64, int64, time.Time, error) {
	row, err := r.queries.GetInviteToken(ctx, token)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	if row.UsedAt.Valid {
		return 0, 0, time.Time{}, domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(row.ExpiresAt) {
		return 0, 0, time.Time{}, domain.ErrTokenExpired
	}

	return row.UserID, row.OwnerID, row.ExpiresAt, nil
}

func (r *Repository) MarkInviteTokenUsed(ctx context.Context, token string) error {
	return r.queries.MarkInviteTokenUsed(ctx, token)
}

func (r *Repository) SetPassword(ctx context.Context, userID int64, passwordHash string) error {
	return r.queries.SetPassword(ctx, sqlc.SetPasswordParams{
		PasswordHash: passwordHash,
		ID:           userID,
	})
}

// GetBusinessByID returns the business name and owner ID.
func (r *Repository) GetBusinessByID(ctx context.Context, businessID int64) (string, int64, error) {
	row, err := r.queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", 0, err
	}
	return row.Name, row.OwnerID, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func permissionFromRow(id, ownerID, userID int64, apps json.RawMessage, role sqlc.NullUserAppPermissionsRole, isActive sql.NullBool, createdAt, updatedAt sql.NullTime) *domain.UserAppPermission {
	p := &domain.UserAppPermission{
		ID:      id,
		OwnerID: ownerID,
		UserID:  userID,
	}

	if role.Valid {
		p.Role = domain.Role(role.UserAppPermissionsRole)
	}

	if isActive.Valid {
		p.IsActive = isActive.Bool
	}

	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}

	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}

	if err := json.Unmarshal(apps, &p.Apps); err != nil {
		p.Apps = []string{}
	}

	return p
}
