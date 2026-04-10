package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	"github.com/voxire/onion/internal/modules/roles/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) ListRoles(ctx context.Context, ownerID int64) ([]domain.Role, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) CreateRole(ctx context.Context, ownerID int64, input domain.CreateRoleInput) (int64, error) {
	if input.Name == "" {
		return 0, errors.New("name is required")
	}

	perms := input.Permissions
	if perms == nil {
		perms = []string{}
	}

	return s.repo.Create(ctx, ownerID, input.Name, perms)
}

func (s *Service) UpdateRole(ctx context.Context, ownerID int64, input domain.UpdateRoleInput) error {
	if input.ID == 0 {
		return errors.New("role id is required")
	}

	perms := input.Permissions
	if perms == nil {
		perms = []string{}
	}

	return s.repo.Update(ctx, ownerID, input.ID, input.Name, perms)
}

func (s *Service) DeleteRole(ctx context.Context, ownerID, roleID int64) error {
	if roleID == 0 {
		return errors.New("role id is required")
	}
	return s.repo.Delete(ctx, ownerID, roleID)
}

func (s *Service) ListPermissions() []domain.PermissionEntry {
	return domain.AllPermissions()
}
