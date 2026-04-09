package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	"github.com/voxire/onion/internal/modules/departments/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListDepartments(ctx context.Context, ownerID int64) ([]domain.Department, error) {
	return s.repo.ListDepartments(ctx, ownerID)
}

func (s *Service) CreateDepartment(ctx context.Context, input domain.DepartmentInput) (int64, error) {
	return s.repo.CreateDepartment(ctx, input)
}

func (s *Service) UpdateDepartment(ctx context.Context, ownerID int64, input domain.DepartmentInput) error {
	return s.repo.UpdateDepartment(ctx, ownerID, input)
}

func (s *Service) DeleteDepartment(ctx context.Context, ownerID, departmentID int64) error {
	return s.repo.DeleteDepartment(ctx, ownerID, departmentID)
}
