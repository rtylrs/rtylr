package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	"github.com/voxire/onion/internal/modules/roles/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Role, error) {
	businessID := ownerID
	rows, err := r.queries.ListRoles(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list roles", logger.Error(err))
		return nil, err
	}

	roles := make([]domain.Role, 0, len(rows))
	for _, row := range rows {
		var role domain.Role
		role.ID = row.ID
		role.OwnerID = row.OwnerID
		role.BusinessID = row.BusinessID
		role.Name = row.Name
		role.IsSystem = row.IsSystem
		if row.CreatedAt.Valid {
			role.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			role.UpdatedAt = row.UpdatedAt.Time
		}
		if err := json.Unmarshal(row.Permissions, &role.Permissions); err != nil {
			role.Permissions = []string{}
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, name string, permissions []string) (int64, error) {
	businessID := ownerID
	permsJSON, err := json.Marshal(permissions)
	if err != nil {
		return 0, err
	}

	res, err := r.queries.CreateRole(ctx, sqlc.CreateRoleParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        name,
		Permissions: permsJSON,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create role", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, roleID int64, name string, permissions []string) error {
	businessID := ownerID
	permsJSON, err := json.Marshal(permissions)
	if err != nil {
		return err
	}

	result, err := r.queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
		Name:        name,
		Permissions: permsJSON,
		ID:          roleID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update role", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, roleID int64) error {
	businessID := ownerID
	result, err := r.queries.DeleteRole(ctx, sqlc.DeleteRoleParams{
		ID:         roleID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete role", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}
