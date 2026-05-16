package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	"github.com/voxire/onion/internal/modules/roles/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	rolesScope = "roles"
	rolesTTL   = 10 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Role, error)
	Create(ctx context.Context, ownerID int64, name string, permissions []string) (int64, error)
	Update(ctx context.Context, ownerID, roleID int64, name string, permissions []string) error
	Delete(ctx context.Context, ownerID, roleID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func invalidateRolesScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(rolesScope, ownerID))
		// Role changes affect every user's effective permissions, so wipe
		// the whole business scope (cached /me, team, etc.).
		cache.DeleteByPattern(ctx, cacheKeys.BusinessPrefix(ownerID))
	}
}

func (s *Service) ListRoles(ctx context.Context, ownerID int64) ([]domain.Role, error) {
	return cacher.Of[[]domain.Role]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(rolesScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Role, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, rolesTTL),
	)
}

func (s *Service) CreateRole(ctx context.Context, ownerID int64, input domain.CreateRoleInput) (int64, error) {
	if input.Name == "" {
		return 0, errors.New("name is required")
	}
	perms := input.Permissions
	if perms == nil {
		perms = []string{}
	}
	id, err := s.repo.Create(ctx, ownerID, input.Name, perms)
	if err != nil {
		return 0, err
	}
	invalidateRolesScope(ctx, ownerID)
	return id, nil
}

func (s *Service) UpdateRole(ctx context.Context, ownerID int64, input domain.UpdateRoleInput) error {
	if input.ID == 0 {
		return errors.New("role id is required")
	}
	perms := input.Permissions
	if perms == nil {
		perms = []string{}
	}
	if err := s.repo.Update(ctx, ownerID, input.ID, input.Name, perms); err != nil {
		return err
	}
	invalidateRolesScope(ctx, ownerID)
	return nil
}

func (s *Service) DeleteRole(ctx context.Context, ownerID, roleID int64) error {
	if roleID == 0 {
		return errors.New("role id is required")
	}
	if err := s.repo.Delete(ctx, ownerID, roleID); err != nil {
		return err
	}
	invalidateRolesScope(ctx, ownerID)
	return nil
}

func (s *Service) ListPermissions() []domain.PermissionEntry {
	return domain.AllPermissions()
}
