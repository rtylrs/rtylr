package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/departments/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	departmentsScope = "departments"
	departmentsTTL   = 5 * time.Minute
)

type repository interface {
	ListDepartments(ctx context.Context, businessID int64) ([]domain.Department, error)
	CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error)
	UpdateDepartment(ctx context.Context, businessID int64, input domain.DepartmentInput) error
	DeleteDepartment(ctx context.Context, businessID, departmentID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateDepartmentsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(departmentsScope, businessID))
}

func (s *Service) ListDepartments(ctx context.Context, businessID int64) ([]domain.Department, error) {
	return cacher.Of[[]domain.Department]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(departmentsScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Department, error) {
			return s.repo.ListDepartments(ctx, businessID)
		},
		cacher.WithExpiry(ctx, departmentsTTL),
	)
}

func (s *Service) CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error) {
	id, err := s.repo.CreateDepartment(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateDepartmentsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) UpdateDepartment(ctx context.Context, businessID int64, input domain.DepartmentInput) error {
	if err := s.repo.UpdateDepartment(ctx, businessID, input); err != nil {
		return err
	}
	invalidateDepartmentsScope(ctx, businessID)
	return nil
}

func (s *Service) DeleteDepartment(ctx context.Context, businessID, departmentID int64) error {
	if err := s.repo.DeleteDepartment(ctx, businessID, departmentID); err != nil {
		return err
	}
	invalidateDepartmentsScope(ctx, businessID)
	return nil
}
