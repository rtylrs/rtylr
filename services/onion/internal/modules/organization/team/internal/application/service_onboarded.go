package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
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
	if input.UserID <= 0 || businessID <= 0 {
		return errors.New("business and user are required")
	}
	ownerID, err := s.requireTeamManager(ctx, actorID, businessID)
	if err != nil {
		return err
	}
	existing, err := s.repo.GetStaffUserByID(ctx, ownerID, businessID, input.UserID)
	if err != nil {
		return err
	}
	if !existing.IsActive || input.UserID == ownerID {
		return domain.ErrForbidden
	}
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, input); err != nil {
		return err
	}

	memberType := input.MemberType
	if memberType == "" {
		memberType = existing.MemberType
	}
	if memberType == "" {
		memberType = "member"
	}
	licensedApps := []string{}
	if input.LicensedAppIDs != nil {
		licensedApps = *input.LicensedAppIDs
	} else {
		licensedApps = existing.LicensedAppIDs
	}
	assignments := make([]domain.AppRoleAssignmentInput, 0, len(existing.PermissionProfileAssignments))
	for _, assignment := range existing.PermissionProfileAssignments {
		if assignment.GroupID != nil || assignment.Source != authorization.SourceExplicit {
			continue
		}
		assignments = append(assignments, domain.AppRoleAssignmentInput{
			AppID: string(assignment.AppID), RoleIDs: append([]int64{}, assignment.RoleIDs...),
			PrimaryRoleID: assignment.PrimaryRoleID, Source: string(assignment.Source),
			Status: "active", Scopes: append([]authorization.Scope{}, assignment.Scopes...),
		})
	}
	if input.PermissionProfileAssignments != nil {
		assignments = *input.PermissionProfileAssignments
	}
	groupIDs := make([]int64, 0, len(existing.Groups))
	for _, group := range existing.Groups {
		groupIDs = append(groupIDs, group.ID)
	}
	if input.GroupIDs != nil {
		groupIDs = *input.GroupIDs
	}
	if input.FullName != nil {
		if err := s.repo.UpdateUserProfile(ctx, input.UserID, *input.FullName); err != nil {
			return err
		}
	}
	if err := s.repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, actorID, input.UserID, licensedApps, memberType, "active", assignments, groupIDs); err != nil {
		return err
	}
	invalidateUserScope(ctx, input.UserID)
	invalidateBusinessTeamScope(ctx, businessID)
	s.recordAudit(ctx, businessID, actorID, "member", input.UserID, "member.access_updated",
		"Updated access for "+memberLabel(existing.FullName, existing.Email))
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
	label := memberLabel("", "")
	if member, lookupErr := s.repo.GetStaffUserByID(ctx, ownerID, businessID, userID); lookupErr == nil && member != nil {
		label = memberLabel(member.FullName, member.Email)
	}

	if err := s.repo.DeactivateUser(ctx, businessID, ownerID, actorID, userID); err != nil {
		return err
	}

	invalidateUserScope(ctx, userID)
	if businessID != 0 {
		invalidateBusinessTeamScope(ctx, businessID)
	}
	s.recordAudit(ctx, businessID, actorID, "member", userID, "member.deactivated", "Deactivated "+label)
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
	if userID <= 0 || ownerID <= 0 {
		return false
	}
	if userID == ownerID {
		return true
	}
	businessID, err := s.repo.GetBusinessIDByOwnerID(ctx, ownerID)
	if err != nil || businessID <= 0 {
		return false
	}
	apps, err := s.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return false
	}
	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, userID, apps)
	return err == nil && found && snapshot.MembershipActive
}

func (s *Service) TeamManagerOwnerID(ctx context.Context, actorID, businessID int64) (int64, error) {
	return s.requireTeamManager(ctx, actorID, businessID)
}

func (s *Service) requireTeamManager(ctx context.Context, actorID, businessID int64) (int64, error) {
	if actorID <= 0 || businessID <= 0 {
		return 0, domain.ErrForbidden
	}
	_, ownerID, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil || ownerID <= 0 {
		return 0, domain.ErrForbidden
	}
	if actorID == ownerID {
		return ownerID, nil
	}
	apps, err := s.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return 0, domain.ErrForbidden
	}
	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps)
	if err != nil || !found || !snapshot.MembershipActive {
		return 0, domain.ErrForbidden
	}
	access := authorization.Resolve(snapshot)
	if access.IsOwner || access.IsAdmin || (containsString(access.DashPermissions, "hub:team:edit") || containsString(access.DashPermissions, "hub:groups:manage")) {
		return ownerID, nil
	}
	return 0, domain.ErrForbidden
}

func (s *Service) enforceAuthorizationDelegation(ctx context.Context, actorID, ownerID, businessID int64, input domain.UpdatePermissionsInput) error {
	if actorID == ownerID {
		return nil
	}
	apps, err := s.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return domain.ErrDelegationCeiling
	}
	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps)
	if err != nil || !found || !snapshot.MembershipActive {
		return domain.ErrDelegationCeiling
	}
	access := authorization.Resolve(snapshot)
	if access.IsOwner || access.IsAdmin {
		return nil
	}
	if input.MemberType == "admin" {
		return domain.ErrDelegationCeiling
	}
	if input.LicensedAppIDs != nil && !containsString(access.DashPermissions, "hub:licenses:manage") {
		return domain.ErrDelegationCeiling
	}
	if input.GroupIDs != nil && !containsString(access.DashPermissions, "hub:team:edit") {
		return domain.ErrDelegationCeiling
	}
	if input.PermissionProfileAssignments == nil || len(*input.PermissionProfileAssignments) == 0 {
		return nil
	}
	if !containsString(access.DashPermissions, "hub:roles:edit") {
		return domain.ErrDelegationCeiling
	}
	for _, assignment := range *input.PermissionProfileAssignments {
		if len(assignment.RoleIDs) == 0 {
			return domain.ErrDelegationCeiling
		}
		for _, profileID := range assignment.RoleIDs {
			grant, grantErr := s.repo.GetAuthorizationRoleGrant(ctx, businessID, profileID)
			if grantErr != nil || grant.AppID != assignment.AppID || !authorization.CanDelegatePermissions(&access, authorization.AppID(grant.AppID), grant.Permissions) {
				return domain.ErrDelegationCeiling
			}
		}
	}
	return nil
}
