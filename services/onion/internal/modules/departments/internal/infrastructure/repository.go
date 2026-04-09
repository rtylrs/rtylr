package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListDepartments(ctx context.Context, ownerID int64) ([]domain.Department, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, description, created_at, updated_at
		 FROM hr_departments WHERE owner_id = ? ORDER BY name ASC`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list departments", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var departments []domain.Department
	for rows.Next() {
		var d domain.Department
		if err := d.FromRow(rows); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}

	return departments, rows.Err()
}

func (r *Repository) CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO hr_departments (owner_id, name, description) VALUES (?, ?, ?)`,
		input.OwnerID, input.Name,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create department", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateDepartment(ctx context.Context, ownerID int64, input domain.DepartmentInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE hr_departments SET name = ?, description = ?, updated_at = NOW() WHERE id = ? AND owner_id = ?`,
		input.Name,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		input.ID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update department", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteDepartment(ctx context.Context, ownerID, departmentID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM hr_departments WHERE id = ? AND owner_id = ?`,
		departmentID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete department", logger.Error(err))
		return err
	}
	return nil
}
