package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	rolesScope = "roles"
)

type normalizedRepository interface {
	ListPermissionCatalog(ctx context.Context) ([]domain.PermissionEntry, error)
	ListNormalized(ctx context.Context, businessID int64, filter domain.RoleFilter) (domain.RolePage, error)
	GetNormalized(ctx context.Context, businessID, roleID int64) (domain.Role, error)
	CreateNormalized(ctx context.Context, businessID, actorID int64, input domain.CreateRoleInput) (domain.Role, error)
	UpdateNormalized(ctx context.Context, businessID, actorID int64, input domain.UpdateRoleInput) (domain.Role, error)
	ArchiveNormalized(ctx context.Context, businessID, actorID, roleID int64, reason string) (domain.Role, error)
	CloneNormalized(ctx context.Context, businessID, actorID, roleID int64, name, description string) (domain.Role, error)
	ListAssignments(ctx context.Context, businessID int64, filter domain.AssignmentFilter) (domain.RoleAssignmentPage, error)
	AssignRole(ctx context.Context, businessID, actorID int64, input domain.AssignRoleInput) (domain.RoleAssignment, error)
	RemoveAssignment(ctx context.Context, businessID, actorID, assignmentID int64) error
	GetAssignment(ctx context.Context, businessID, assignmentID int64) (domain.RoleAssignment, error)
	PreviewPermissionImpact(ctx context.Context, businessID, roleID int64, permissions []string) (domain.PermissionImpact, error)
	ListAudit(ctx context.Context, businessID int64, filter domain.AuditFilter) (domain.AuditPage, error)
}

type Service struct {
	normalizedRepo normalizedRepository
}

func NewService(_ context.Context, repo normalizedRepository) *Service {
	return &Service{
		normalizedRepo: repo,
	}
}

func invalidateRolesScope(ctx context.Context, businessID int64) {
	// Role changes affect every user's effective permissions, so wipe the whole
	// business scope. The cached /me payload holds EffectiveAccess and is keyed
	// per user, so it needs its own pattern: BusinessPrefix does not match it.
	cacher.InvalidateScope(ctx,
		cacheKeys.ScopedPrefix(rolesScope, businessID),
		cacheKeys.BusinessPrefix(businessID),
		cacheKeys.BusinessUserMePrefix(businessID),
	)
}

func (s *Service) ListPermissionCatalog(ctx context.Context) ([]domain.PermissionEntry, error) {
	if s.normalizedRepo == nil {
		return nil, domain.ErrAuthorizationSchemaUnavailable
	}
	return s.normalizedRepo.ListPermissionCatalog(ctx)
}

func (s *Service) ListManagedRoles(ctx context.Context, businessID int64, filter domain.RoleFilter) (domain.RolePage, error) {
	filter = normalizeRoleFilter(filter)
	if s.normalizedRepo == nil {
		return domain.RolePage{}, domain.ErrAuthorizationSchemaUnavailable
	}
	return s.normalizedRepo.ListNormalized(ctx, businessID, filter)
}

func (s *Service) GetManagedRole(ctx context.Context, businessID, roleID int64) (domain.Role, error) {
	if s.normalizedRepo == nil {
		return domain.Role{}, domain.ErrAuthorizationSchemaUnavailable
	}
	return s.normalizedRepo.GetNormalized(ctx, businessID, roleID)
}

func (s *Service) CreateManagedRole(ctx context.Context, businessID, actorID int64, input domain.CreateRoleInput) (domain.Role, error) {
	if s.normalizedRepo == nil {
		return domain.Role{}, domain.ErrAuthorizationSchemaUnavailable
	}
	if input.Name == "" {
		return domain.Role{}, errors.New("name is required")
	}
	if input.AppID == "" {
		input.AppID = "dash"
	}
	input.Permissions = nonNilStrings(input.Permissions)
	input.ScopeRules = nonNilStrings(input.ScopeRules)
	role, err := s.normalizedRepo.CreateNormalized(ctx, businessID, actorID, input)
	if err == nil {
		invalidateRolesScope(ctx, businessID)
	}
	return role, err
}

func (s *Service) UpdateManagedRole(ctx context.Context, businessID, actorID int64, input domain.UpdateRoleInput) (domain.Role, error) {
	if s.normalizedRepo == nil {
		return domain.Role{}, domain.ErrAuthorizationSchemaUnavailable
	}
	if input.ID <= 0 {
		return domain.Role{}, errors.New("role id is required")
	}
	if input.Name == "" {
		return domain.Role{}, errors.New("name is required")
	}
	input.Permissions = nonNilStrings(input.Permissions)
	input.ScopeRules = nonNilStrings(input.ScopeRules)
	role, err := s.normalizedRepo.UpdateNormalized(ctx, businessID, actorID, input)
	if err == nil {
		invalidateRolesScope(ctx, businessID)
	}
	return role, err
}

func (s *Service) ArchiveManagedRole(ctx context.Context, businessID, actorID, roleID int64, reason string) (domain.Role, error) {
	if s.normalizedRepo == nil {
		return domain.Role{}, domain.ErrAuthorizationSchemaUnavailable
	}
	if roleID <= 0 {
		return domain.Role{}, errors.New("role id is required")
	}
	role, err := s.normalizedRepo.ArchiveNormalized(ctx, businessID, actorID, roleID, reason)
	if err == nil {
		invalidateRolesScope(ctx, businessID)
	}
	return role, err
}

func (s *Service) CloneManagedRole(ctx context.Context, businessID, actorID, roleID int64, name, description string) (domain.Role, error) {
	if s.normalizedRepo == nil {
		return domain.Role{}, domain.ErrAuthorizationSchemaUnavailable
	}
	if roleID <= 0 || name == "" {
		return domain.Role{}, errors.New("role id and name are required")
	}
	role, err := s.normalizedRepo.CloneNormalized(ctx, businessID, actorID, roleID, name, description)
	if err == nil {
		invalidateRolesScope(ctx, businessID)
	}
	return role, err
}

func (s *Service) ListManagedAssignments(ctx context.Context, businessID int64, filter domain.AssignmentFilter) (domain.RoleAssignmentPage, error) {
	if s.normalizedRepo == nil {
		return domain.RoleAssignmentPage{}, domain.ErrAuthorizationSchemaUnavailable
	}
	filter = normalizeAssignmentFilter(filter)
	return s.normalizedRepo.ListAssignments(ctx, businessID, filter)
}

func (s *Service) AssignManagedRole(ctx context.Context, businessID, actorID int64, input domain.AssignRoleInput) (domain.RoleAssignment, error) {
	if input.RoleID <= 0 || input.AppID == "" {
		return domain.RoleAssignment{}, errors.New("role id and app id are required")
	}
	if (input.UserID == nil) == (input.GroupID == nil) {
		return domain.RoleAssignment{}, domain.ErrInvalidAssignmentTarget
	}
	if s.normalizedRepo == nil {
		return domain.RoleAssignment{}, domain.ErrAuthorizationSchemaUnavailable
	}
	input.Scopes = nonNilStrings(input.Scopes)
	assignment, err := s.normalizedRepo.AssignRole(ctx, businessID, actorID, input)
	if err == nil {
		invalidateRolesScope(ctx, businessID)
	}
	return assignment, err
}

func (s *Service) GetManagedAssignment(ctx context.Context, businessID, assignmentID int64) (domain.RoleAssignment, error) {
	if s.normalizedRepo == nil {
		return domain.RoleAssignment{}, domain.ErrAuthorizationSchemaUnavailable
	}
	if assignmentID <= 0 {
		return domain.RoleAssignment{}, domain.ErrRoleNotFound
	}
	return s.normalizedRepo.GetAssignment(ctx, businessID, assignmentID)
}

func (s *Service) RemoveManagedAssignment(ctx context.Context, businessID, actorID, assignmentID int64) error {
	if s.normalizedRepo == nil {
		return domain.ErrAuthorizationSchemaUnavailable
	}
	if assignmentID <= 0 {
		return errors.New("assignment id is required")
	}
	err := s.normalizedRepo.RemoveAssignment(ctx, businessID, actorID, assignmentID)
	if err == nil {
		invalidateRolesScope(ctx, businessID)
	}
	return err
}

func (s *Service) PreviewManagedPermissionImpact(ctx context.Context, businessID, roleID int64, permissions []string) (domain.PermissionImpact, error) {
	if s.normalizedRepo == nil {
		return domain.PermissionImpact{}, domain.ErrAuthorizationSchemaUnavailable
	}
	return s.normalizedRepo.PreviewPermissionImpact(ctx, businessID, roleID, nonNilStrings(permissions))
}

func (s *Service) ListManagedAudit(ctx context.Context, businessID int64, filter domain.AuditFilter) (domain.AuditPage, error) {
	if s.normalizedRepo == nil {
		return domain.AuditPage{}, domain.ErrAuthorizationSchemaUnavailable
	}
	filter = normalizeAuditFilter(filter)
	return s.normalizedRepo.ListAudit(ctx, businessID, filter)
}

func normalizeRoleFilter(filter domain.RoleFilter) domain.RoleFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return filter
}

func normalizeAssignmentFilter(filter domain.AssignmentFilter) domain.AssignmentFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return filter
}

func normalizeAuditFilter(filter domain.AuditFilter) domain.AuditFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return filter
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
