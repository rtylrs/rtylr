package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
)

func (s *Service) MarkOnboarded(ctx context.Context, userID int64) error {
	if err := s.repo.MarkOnboarded(ctx, userID); err != nil {
		return err
	}
	invalidateUserScope(ctx, userID)
	return nil
}

// GetBusinessInfo returns the business name, owner ID, and whether the user is the owner.
func (s *Service) GetBusinessInfo(ctx context.Context, businessID, userID int64) (string, int64, bool) {
	if businessID == 0 {
		return "", 0, false
	}

	name, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", 0, false
	}

	return name, ownerID, ownerID == userID
}

func (s *Service) UpdatePermissions(ctx context.Context, actorID, businessID int64, input domain.UpdatePermissionsInput) error {
	if input.UserID == 0 {
		return errors.New("user id is required")
	}

	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return err
	}
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, input); err != nil {
		return err
	}

	role := input.Role
	apps := []string{}
	if input.Apps != nil {
		apps = *input.Apps
	}
	if input.Apps == nil || role == "" {
		existing, err := s.repo.GetStaffUserByID(ctx, ownerID, businessID, input.UserID)
		if err != nil {
			return err
		}
		if input.Apps == nil {
			apps = existing.Apps
		}
		if role == "" {
			role = existing.Role
		}
	}
	if len(apps) == 0 && len(input.AppAssignments) > 0 {
		for _, assignment := range input.AppAssignments {
			if assignment.AppID != "" {
				apps = appendUniqueString(apps, assignment.AppID)
			}
		}
	}
	if role == "" {
		role = domain.RoleStaff
	}
	if !isAssignableRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}
	if role == domain.RoleOwner && input.UserID != ownerID {
		return domain.ErrForbidden
	}

	if input.FullName != nil {
		if err := s.repo.UpdateUserProfile(ctx, input.UserID, *input.FullName); err != nil {
			return err
		}
	}

	if err := s.repo.UpsertPermission(ctx, businessID, ownerID, input.UserID, apps, role); err != nil {
		return err
	}
	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, input.UserID, apps, role, input.DashRole, input.AppAssignments, input.GroupIDs); err != nil {
		return err
	}

	invalidateUserScope(ctx, input.UserID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return nil
}

func (s *Service) DeactivateUser(ctx context.Context, actorID, businessID, userID int64) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return err
	}
	if userID == ownerID {
		return domain.ErrForbidden
	}

	if err := s.repo.DeactivateUser(ctx, ownerID, userID); err != nil {
		return err
	}

	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	return nil
}

func (s *Service) resolveBusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	if businessID == 0 {
		return fallbackOwnerID, nil
	}

	_, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return 0, err
	}

	return ownerID, nil
}

func (s *Service) BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	return s.resolveBusinessOwnerID(ctx, fallbackOwnerID, businessID)
}

func (s *Service) CanAccessOwner(ctx context.Context, userID, ownerID int64) bool {
	if userID == 0 || ownerID == 0 {
		return false
	}
	if userID == ownerID {
		return true
	}
	if businessID, businessErr := s.repo.GetBusinessIDByOwnerID(ctx, ownerID); businessErr == nil && businessID > 0 {
		apps, _ := s.GetOwnerApps(ctx, ownerID)
		if snapshot, found, snapshotErr := s.repo.GetAuthorizationSnapshot(ctx, businessID, userID, apps); snapshotErr == nil && found {
			return authorization.Resolve(snapshot).Apps != nil && snapshot.MembershipActive
		}
	}
	perm, err := s.repo.GetPermission(ctx, ownerID, userID)
	if err != nil {
		return false
	}
	return perm != nil && perm.IsActive
}

func (s *Service) TeamManagerOwnerID(ctx context.Context, actorID, businessID int64) (int64, error) {
	return s.requireTeamManager(ctx, actorID, businessID)
}

func (s *Service) requireTeamManager(ctx context.Context, actorID, businessID int64) (int64, error) {
	ownerID, err := s.resolveBusinessOwnerID(ctx, actorID, businessID)
	if err != nil {
		return 0, err
	}
	if actorID == ownerID {
		return ownerID, nil
	}
	if businessID == 0 {
		return 0, domain.ErrForbidden
	}
	if apps, appsErr := s.GetOwnerApps(ctx, ownerID); appsErr == nil {
		if snapshot, found, snapshotErr := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps); snapshotErr == nil && found {
			resolved := authorization.Resolve(snapshot)
			if resolved.IsOwner || resolved.IsAdmin || containsString(resolved.DashPermissions, "hub:team:edit") {
				return ownerID, nil
			}
			return 0, domain.ErrForbidden
		}
	}

	perm, err := s.repo.GetPermissionByBusiness(ctx, businessID, actorID)
	if err != nil {
		return 0, domain.ErrForbidden
	}
	if perm != nil && perm.IsActive && (perm.Role == domain.RoleAdmin || perm.Role == domain.RoleOwner) {
		return ownerID, nil
	}

	return 0, domain.ErrForbidden
}

func (s *Service) enforceAuthorizationDelegation(ctx context.Context, actorID, ownerID, businessID int64, input domain.UpdatePermissionsInput) error {
	if actorID == ownerID {
		return nil
	}
	if businessID <= 0 {
		return domain.ErrForbidden
	}

	apps, err := s.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return domain.ErrForbidden
	}
	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps)
	if err != nil || !found {
		// Do not allow a legacy fallback actor to submit normalized role,
		// license, group, or Dash-role fields during the migration window.
		if input.DashRole != "" || input.Apps != nil || input.Role != "" || input.AppAssignments != nil || input.GroupIDs != nil {
			return domain.ErrDelegationCeiling
		}
		return nil
	}
	access := authorization.Resolve(snapshot)
	if access.IsOwner || access.IsAdmin {
		return nil
	}

	hasRoleEdit := containsString(access.DashPermissions, "hub:roles:edit")
	hasTeamEdit := containsString(access.DashPermissions, "hub:team:edit")
	hasLicenseManage := containsString(access.DashPermissions, "hub:licenses:manage")
	if input.DashRole != "" {
		if !hasRoleEdit || !authorization.CanAssignDashboardRole(&access, input.DashRole) {
			return domain.ErrDelegationCeiling
		}
	}
	if input.Role != "" && !hasRoleEdit {
		return domain.ErrDelegationCeiling
	}
	if input.Apps != nil && !hasLicenseManage {
		return domain.ErrDelegationCeiling
	}
	if input.GroupIDs != nil && !hasTeamEdit {
		return domain.ErrDelegationCeiling
	}
	if len(input.AppAssignments) == 0 {
		return nil
	}
	if !hasRoleEdit {
		return domain.ErrDelegationCeiling
	}
	for _, assignment := range input.AppAssignments {
		if assignment.Source == string(authorization.SourceDashDefault) || assignment.Source == string(authorization.SourceSystem) {
			return domain.ErrDelegationCeiling
		}
		if len(assignment.RoleIDs) == 0 {
			return domain.ErrDelegationCeiling
		}
		for _, roleID := range assignment.RoleIDs {
			grant, grantErr := s.repo.GetAuthorizationRoleGrant(ctx, businessID, roleID)
			if grantErr != nil || grant.AppID != assignment.AppID || !authorization.CanDelegatePermissions(&access, authorization.AppID(grant.AppID), grant.Permissions) {
				return domain.ErrDelegationCeiling
			}
		}
	}
	return nil
}
