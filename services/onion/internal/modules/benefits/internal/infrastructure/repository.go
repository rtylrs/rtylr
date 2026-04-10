package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.BenefitPlan, error) {
	query := `
		SELECT id, owner_id, name, type, provider, cost_employer, cost_employee, is_active, created_at
		FROM hr_benefit_plans
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list benefit plans", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.BenefitPlan
	for rows.Next() {
		var item domain.BenefitPlan
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.BenefitPlanInput) (int64, error) {
	query := `
		INSERT INTO hr_benefit_plans (owner_id, name, type, provider, cost_employer, cost_employee, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	benefitType := input.Type
	if benefitType == "" {
		benefitType = "other"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.Name, benefitType,
		sql.NullString{String: input.Provider, Valid: input.Provider != ""},
		input.CostEmployer, input.CostEmployee, input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create benefit plan", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}
