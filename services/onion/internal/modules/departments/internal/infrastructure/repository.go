package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	"github.com/voxire/onion/internal/modules/departments/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListDepartments(ctx context.Context, businessID int64) ([]domain.Department, error) {
	rows, err := r.queries.ListDepartments(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list departments", logger.Error(err))
		return nil, err
	}

	departments := make([]domain.Department, 0, len(rows))
	for _, row := range rows {
		var d domain.Department
		d.ID = row.ID
		d.OwnerID = row.OwnerID
		d.BusinessID = row.BusinessID
		d.Name = row.Name
		if row.Description.Valid {
			d.Description = row.Description.String
		}
		if row.ManagerStaffID.Valid {
			d.ManagerStaffID = row.ManagerStaffID.Int64
		}
		if row.CreatedAt.Valid {
			d.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			d.UpdatedAt = row.UpdatedAt.Time
		}
		departments = append(departments, d)
	}

	return departments, nil
}

func (r *Repository) CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error) {
	result, err := r.queries.CreateDepartment(ctx, sqlc.CreateDepartmentParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create department", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateDepartment(ctx context.Context, businessID int64, input domain.DepartmentInput) error {
	err := r.queries.UpdateDepartment(ctx, sqlc.UpdateDepartmentParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		ID:          input.ID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update department", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteDepartment(ctx context.Context, businessID, departmentID int64) error {
	err := r.queries.DeleteDepartment(ctx, sqlc.DeleteDepartmentParams{
		ID:         departmentID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete department", logger.Error(err))
		return err
	}
	return nil
}
