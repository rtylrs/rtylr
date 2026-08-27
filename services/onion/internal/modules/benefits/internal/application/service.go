package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	"github.com/voxire/onion/internal/modules/benefits/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	benefitsScope = "benefits"
	benefitsTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.BenefitPlan, error)
	Create(ctx context.Context, businessID int64, input domain.BenefitPlanInput) (int64, error)
	CreateEnrollment(ctx context.Context, businessID, employeeID, planID int64) (domain.BenefitEnrollment, error)
	ListEnrollmentsByEmployee(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.BenefitEnrollment, int32, error)
}

type Service struct {
	repo      repository
	employees domain.EmployeeDirectory
}

func NewService(_ context.Context, db *sql.DB, employees domain.EmployeeDirectory) *Service {
	return &Service{
		repo:      infrastructure.NewRepository(db),
		employees: employees,
	}
}

func invalidateBenefitsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(benefitsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.BenefitPlan, error) {
	return cacher.Of[[]domain.BenefitPlan]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(benefitsScope, businessID, "list"),
		func(ctx context.Context) ([]domain.BenefitPlan, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, benefitsTTL),
	)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.BenefitPlanInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateBenefitsScope(ctx, businessID)
	return id, nil
}
