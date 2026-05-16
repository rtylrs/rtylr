package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	"github.com/voxire/onion/internal/modules/departments/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	departmentsScope = "departments"
	departmentsTTL   = 5 * time.Minute
)

type repository interface {
	ListDepartments(ctx context.Context, ownerID int64) ([]domain.Department, error)
	CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error)
	UpdateDepartment(ctx context.Context, ownerID int64, input domain.DepartmentInput) error
	DeleteDepartment(ctx context.Context, ownerID, departmentID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidateDepartmentsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(departmentsScope, ownerID))
	}
}

func (s *Service) ListDepartments(ctx context.Context, ownerID int64) ([]domain.Department, error) {
	return cacher.Of[[]domain.Department]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(departmentsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Department, error) {
			return s.repo.ListDepartments(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, departmentsTTL),
	)
}

func (s *Service) CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error) {
	id, err := s.repo.CreateDepartment(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateDepartmentsScope(ctx, input.OwnerID)
	return id, nil
}

func (s *Service) UpdateDepartment(ctx context.Context, ownerID int64, input domain.DepartmentInput) error {
	if err := s.repo.UpdateDepartment(ctx, ownerID, input); err != nil {
		return err
	}
	invalidateDepartmentsScope(ctx, ownerID)
	return nil
}

func (s *Service) DeleteDepartment(ctx context.Context, ownerID, departmentID int64) error {
	if err := s.repo.DeleteDepartment(ctx, ownerID, departmentID); err != nil {
		return err
	}
	invalidateDepartmentsScope(ctx, ownerID)
	return nil
}
