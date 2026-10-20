package application

import (
	"context"
	"errors"
	"sort"
	"strings"

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
	currentGroupIDs := make([]int64, 0, len(existing.Groups))
	for _, group := range existing.Groups {
		currentGroupIDs = append(currentGroupIDs, group.ID)
	}
	if err := s.enforceAuthorizationDelegation(ctx, actorID, ownerID, businessID, input, &delegationCurrent{
		MemberType: existing.MemberType,
		Grants:     grantsOfAssignments(existing.PermissionProfileAssignments),
		GroupIDs:   currentGroupIDs,
	}); err != nil {
		return err
	}

	memberType := input.MemberType
	if memberType == "" {
		memberType = existing.MemberType
	}
	if memberType == "" {
		memberType = "member"
	}
	licensedApps := existing.LicensedAppIDs
	if input.LicensedAppIDs != nil {
		licensedApps = *input.LicensedAppIDs
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

// delegationCurrent is what the target holds before an edit, so the ceiling
// applies to what the edit takes away as well as to what it grants. nil means
// the target is new.
type delegationCurrent struct {
	MemberType string
	Grants     []roleGrant
	GroupIDs   []int64
}

type roleGrantReader interface {
	GetAuthorizationRoleGrant(ctx context.Context, businessID, roleID int64) (domain.AuthorizationRoleGrant, error)
}

type delegationStore interface {
	roleGrantReader
	GetGroup(ctx context.Context, ownerID, businessID, groupID int64) (*domain.Group, error)
}

// roleGrant is one role granted in one app with its scopes.
type roleGrant struct {
	AppID  string
	RoleID int64
	Scopes string
}

func grantsOfAssignments(assignments []authorization.RoleAssignment) []roleGrant {
	var grants []roleGrant
	for _, assignment := range assignments {
		if assignment.GroupID != nil || assignment.Source != authorization.SourceExplicit {
			continue
		}
		for _, roleID := range assignment.RoleIDs {
			grants = append(grants, roleGrant{AppID: string(assignment.AppID), RoleID: roleID, Scopes: scopeSignature(assignment.Scopes)})
		}
	}
	return grants
}

func grantsOfInputs(inputs []domain.AppRoleAssignmentInput) []roleGrant {
	var grants []roleGrant
	for _, input := range inputs {
		for _, roleID := range input.RoleIDs {
			grants = append(grants, roleGrant{AppID: input.AppID, RoleID: roleID, Scopes: scopeSignature(input.Scopes)})
		}
	}
	return grants
}

func scopeSignature(scopes []authorization.Scope) string {
	parts := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		parts = append(parts, scope.Type+"="+scope.ID)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func diffGrants(current, next []roleGrant) (added, removed []roleGrant) {
	have := make(map[roleGrant]struct{}, len(current))
	for _, grant := range current {
		have[grant] = struct{}{}
	}
	want := make(map[roleGrant]struct{}, len(next))
	for _, grant := range next {
		want[grant] = struct{}{}
		if _, ok := have[grant]; !ok {
			added = append(added, grant)
		}
	}
	for _, grant := range current {
		if _, ok := want[grant]; !ok {
			removed = append(removed, grant)
		}
	}
	return added, removed
}

// grantsWithinCeiling reports whether the actor may hand out, or take away,
// every listed role grant.
func grantsWithinCeiling(ctx context.Context, reader roleGrantReader, businessID int64, access *authorization.EffectiveAccess, grants []roleGrant) bool {
	for _, grant := range grants {
		role, err := reader.GetAuthorizationRoleGrant(ctx, businessID, grant.RoleID)
		if err != nil || role.AppID != grant.AppID || !authorization.CanDelegatePermissions(access, authorization.AppID(role.AppID), role.Permissions) {
			return false
		}
	}
	return true
}

// delegationAccess resolves the actor's access. unrestricted is true for the
// owner and business admins, who are not subject to a ceiling.
func (s *Service) delegationAccess(ctx context.Context, actorID, ownerID, businessID int64) (access authorization.EffectiveAccess, unrestricted bool, err error) {
	if actorID == ownerID {
		return authorization.EffectiveAccess{}, true, nil
	}
	apps, err := s.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return authorization.EffectiveAccess{}, false, domain.ErrDelegationCeiling
	}
	snapshot, found, err := s.repo.GetAuthorizationSnapshot(ctx, businessID, actorID, apps)
	if err != nil || !found || !snapshot.MembershipActive {
		return authorization.EffectiveAccess{}, false, domain.ErrDelegationCeiling
	}
	access = authorization.Resolve(snapshot)
	return access, access.IsOwner || access.IsAdmin, nil
}

// enforceGroupMembershipDelegation applies the ceiling to adding someone to, or
// removing someone from, a group: membership confers or strips the group's roles.
func (s *Service) enforceGroupMembershipDelegation(ctx context.Context, actorID, ownerID, businessID int64, group *domain.Group) error {
	access, unrestricted, err := s.delegationAccess(ctx, actorID, ownerID, businessID)
	if err != nil {
		return err
	}
	if unrestricted || group == nil {
		return nil
	}
	if !grantsWithinCeiling(ctx, s.repo, businessID, &access, grantsOfGroupAssignments(group.PermissionProfileAssignments)) {
		return domain.ErrDelegationCeiling
	}
	return nil
}

func grantsOfGroupAssignments(assignments []authorization.RoleAssignment) []roleGrant {
	var grants []roleGrant
	for _, assignment := range assignments {
		for _, roleID := range assignment.RoleIDs {
			grants = append(grants, roleGrant{AppID: string(assignment.AppID), RoleID: roleID, Scopes: scopeSignature(assignment.Scopes)})
		}
	}
	return grants
}

func (s *Service) enforceAuthorizationDelegation(ctx context.Context, actorID, ownerID, businessID int64, input domain.UpdatePermissionsInput, current *delegationCurrent) error {
	access, unrestricted, err := s.delegationAccess(ctx, actorID, ownerID, businessID)
	if err != nil {
		return err
	}
	if unrestricted {
		return nil
	}
	return checkDelegation(ctx, s.repo, &access, ownerID, businessID, input, current)
}

// checkDelegation applies the ceiling of a non-admin actor to an edit.
func checkDelegation(ctx context.Context, store delegationStore, access *authorization.EffectiveAccess, ownerID, businessID int64, input domain.UpdatePermissionsInput, current *delegationCurrent) error {
	if input.MemberType == "admin" || (current != nil && current.MemberType == "admin" && input.MemberType != "" && input.MemberType != "admin") {
		return domain.ErrDelegationCeiling
	}
	if input.LicensedAppIDs != nil && !containsString(access.DashPermissions, "hub:licenses:manage") {
		return domain.ErrDelegationCeiling
	}
	if input.GroupIDs != nil {
		if !containsString(access.DashPermissions, "hub:team:edit") {
			return domain.ErrDelegationCeiling
		}
		var currentGroups []int64
		if current != nil {
			currentGroups = current.GroupIDs
		}
		for _, groupID := range changedGroupIDs(currentGroups, *input.GroupIDs) {
			group, groupErr := store.GetGroup(ctx, ownerID, businessID, groupID)
			if groupErr != nil || group == nil || !grantsWithinCeiling(ctx, store, businessID, access, grantsOfGroupAssignments(group.PermissionProfileAssignments)) {
				return domain.ErrDelegationCeiling
			}
		}
	}
	if input.PermissionProfileAssignments == nil {
		return nil
	}
	for _, assignment := range *input.PermissionProfileAssignments {
		if len(assignment.RoleIDs) == 0 {
			return domain.ErrDelegationCeiling
		}
	}
	var currentGrants []roleGrant
	if current != nil {
		currentGrants = current.Grants
	}
	added, removed := diffGrants(currentGrants, grantsOfInputs(*input.PermissionProfileAssignments))
	if len(added) == 0 && len(removed) == 0 {
		return nil
	}
	if !containsString(access.DashPermissions, "hub:roles:edit") {
		return domain.ErrDelegationCeiling
	}
	if !grantsWithinCeiling(ctx, store, businessID, access, added) || !grantsWithinCeiling(ctx, store, businessID, access, removed) {
		return domain.ErrDelegationCeiling
	}
	return nil
}

func changedGroupIDs(current, next []int64) []int64 {
	seen := make(map[int64]int, len(current)+len(next))
	for _, id := range current {
		seen[id] |= 1
	}
	for _, id := range next {
		seen[id] |= 2
	}
	var changed []int64
	for id, mask := range seen {
		if mask != 3 {
			changed = append(changed, id)
		}
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i] < changed[j] })
	return changed
}
