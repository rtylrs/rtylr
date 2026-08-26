package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	"github.com/voxire/onion/internal/modules/benefits/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.BenefitPlan, error) {
	rows, err := r.queries.ListBenefitPlans(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list benefit plans", logger.Error(err))
		return nil, err
	}

	items := make([]domain.BenefitPlan, 0, len(rows))
	for _, row := range rows {
		var item domain.BenefitPlan
		item.ID = row.ID
		item.OwnerID = row.OwnerID
		item.BusinessID = row.BusinessID
		item.Name = row.Name
		if row.Type.Valid {
			item.Type = string(row.Type.HrBenefitPlansType)
		}
		if row.Provider.Valid {
			item.Provider = row.Provider.String
		}
		if row.CostEmployer.Valid {
			item.CostEmployer = sqlutil.ParseDecimal(row.CostEmployer.String)
		}
		if row.CostEmployee.Valid {
			item.CostEmployee = sqlutil.ParseDecimal(row.CostEmployee.String)
		}
		if row.IsActive.Valid {
			item.IsActive = row.IsActive.Bool
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.BenefitPlanInput) (int64, error) {
	benefitType := input.Type
	if benefitType == "" {
		benefitType = "other"
	}

	res, err := r.queries.CreateBenefitPlan(ctx, sqlc.CreateBenefitPlanParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		Name:         input.Name,
		Type:         sqlc.NullHrBenefitPlansType{HrBenefitPlansType: sqlc.HrBenefitPlansType(benefitType), Valid: true},
		Provider:     sql.NullString{String: input.Provider, Valid: input.Provider != ""},
		CostEmployer: sql.NullString{String: strconv.FormatFloat(input.CostEmployer, 'f', 2, 64), Valid: true},
		CostEmployee: sql.NullString{String: strconv.FormatFloat(input.CostEmployee, 'f', 2, 64), Valid: true},
		IsActive:     sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create benefit plan", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
