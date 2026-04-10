package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Role, error) {
	query := `
		SELECT id, owner_id, name, permissions, is_system, created_at, updated_at
		FROM hub_roles
		WHERE owner_id = ?
		ORDER BY is_system DESC, name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list roles", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		var permsJSON []byte
		if err := rows.Scan(&role.ID, &role.OwnerID, &role.Name, &permsJSON, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan role row", logger.Error(err))
			return nil, err
		}
		if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
			role.Permissions = []string{}
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, name string, permissions []string) (int64, error) {
	permsJSON, err := json.Marshal(permissions)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO hub_roles (owner_id, name, permissions) VALUES (?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query, ownerID, name, permsJSON)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create role", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, roleID int64, name string, permissions []string) error {
	permsJSON, err := json.Marshal(permissions)
	if err != nil {
		return err
	}

	query := `
		UPDATE hub_roles
		SET name = ?, permissions = ?, updated_at = NOW()
		WHERE id = ? AND owner_id = ? AND is_system = FALSE`

	_, err = r.pool.ExecContext(ctx, query, name, permsJSON, roleID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update role", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, roleID int64) error {
	query := `DELETE FROM hub_roles WHERE id = ? AND owner_id = ? AND is_system = FALSE`

	_, err := r.pool.ExecContext(ctx, query, roleID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete role", logger.Error(err))
		return err
	}

	return nil
}
