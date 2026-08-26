package application

import (
	"context"
	"database/sql"
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

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func invalidateRolesScope(ctx context.Context, businessID int64) {
	// Role changes affect every user's effective permissions, so wipe
	// the whole business scope (cached /me, team, etc.).
	cacher.InvalidateScope(ctx,
		cacheKeys.ScopedPrefix(rolesScope, businessID),
		cacheKeys.BusinessPrefix(businessID),
	)
}

func (s *Service) ListRoles(ctx context.Context, ownerID int64) ([]domain.Role, error) {
	businessID := ownerID
	return cacher.Of[[]domain.Role]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(rolesScope, businessID, "list"),
		func(ctx context.Context) ([]domain.Role, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, rolesTTL),
	)
}

func (s *Service) CreateRole(ctx context.Context, ownerID int64, input domain.CreateRoleInput) (int64, error) {
	businessID := ownerID
	if input.Name == "" {
		return 0, errors.New("name is required")
	}
	perms := input.Permissions
	if perms == nil {
		perms = []string{}
	}
	id, err := s.repo.Create(ctx, businessID, input.Name, perms)
	if err != nil {
		return 0, err
	}
	invalidateRolesScope(ctx, businessID)
	return id, nil
}

func (s *Service) UpdateRole(ctx context.Context, ownerID int64, input domain.UpdateRoleInput) error {
	businessID := ownerID
	if input.ID == 0 {
		return errors.New("role id is required")
	}
	perms := input.Permissions
	if perms == nil {
		perms = []string{}
	}
	if err := s.repo.Update(ctx, businessID, input.ID, input.Name, perms); err != nil {
		return err
	}
	invalidateRolesScope(ctx, businessID)
	return nil
}

func (s *Service) DeleteRole(ctx context.Context, ownerID, roleID int64) error {
	businessID := ownerID
	if roleID == 0 {
		return errors.New("role id is required")
	}
	if err := s.repo.Delete(ctx, businessID, roleID); err != nil {
		return err
	}
	invalidateRolesScope(ctx, businessID)
	return nil
}

func (s *Service) ListPermissions() []domain.PermissionEntry {
	return domain.AllPermissions()
}
