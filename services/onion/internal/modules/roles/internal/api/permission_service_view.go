package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

func canViewRoles(user *authPkg.AuthUser) bool {
	if user == nil {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		return authPkg.IsBusinessAdmin(user)
	}
	if user.EffectiveAccess != nil {
		return authPkg.HasDashboardPermission(user, "hub:roles:view")
	}
	return authPkg.IsBusinessAdmin(user)
}

func canEditRoles(user *authPkg.AuthUser) bool {
	if user == nil {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		return authPkg.IsBusinessAdmin(user)
	}
	if user.EffectiveAccess != nil {
		return authPkg.HasDashboardPermission(user, "hub:roles:create") ||
			authPkg.HasDashboardPermission(user, "hub:roles:edit")
	}
	return authPkg.IsBusinessAdmin(user)
}

func canArchiveRoles(user *authPkg.AuthUser) bool {
	if user == nil {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		return authPkg.IsBusinessAdmin(user)
	}
	if user.EffectiveAccess != nil {
		return authPkg.HasDashboardPermission(user, "hub:roles:archive")
	}
	return authPkg.IsBusinessAdmin(user)
}

func canDelegateRoleMutation(user *authPkg.AuthUser, appID string, permissions []string) bool {
	if user == nil {
		return false
	}
	if user.EffectiveAccess == nil {
		return authPkg.IsBusinessAdmin(user)
	}
	return authorization.CanDelegatePermissions(user.EffectiveAccess, authorization.AppID(appID), permissions)
}

func roleAPIError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrRoleNotFound):
		return onekiterrors.New(http.StatusNotFound, "Role not found")
	case errors.Is(err, domain.ErrSystemRoleMutation), errors.Is(err, domain.ErrRoleArchived):
		return onekiterrors.New(http.StatusConflict, "Role cannot be changed")
	case errors.Is(err, domain.ErrInvalidPermission), errors.Is(err, domain.ErrInvalidAssignmentTarget), errors.Is(err, domain.ErrInvalidScope):
		return onekiterrors.New(http.StatusBadRequest, "Invalid role authorization input")
	case errors.Is(err, domain.ErrDelegationCeiling):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrAuthorizationSchemaUnavailable):
		return onekiterrors.New(http.StatusInternalServerError, "Role authorization is not available")
	default:
		return onekiterrors.New(http.StatusInternalServerError, fallback)
	}
}

func nextPage(page, pageSize, total int64) *int32 {
	if page <= 0 || pageSize <= 0 || page*pageSize >= total {
		return nil
	}
	next := int32(page + 1) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	return &next
}

func buildPermissionMatrix(entries []domain.PermissionEntry, roles []domain.Role) *hubpermissionv1.PermissionMatrix {
	roleNames := make([]string, 0, len(roles))
	rolePermSets := make(map[string]map[string]struct{}, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		perms := make(map[string]struct{}, len(role.Permissions))
		for _, p := range role.Permissions {
			perms[p] = struct{}{}
		}
		rolePermSets[role.Name] = perms
	}

	rows := make([]*hubpermissionv1.PermissionRow, 0, len(entries))
	for _, entry := range entries {
		roleValues := make(map[string]bool, len(roleNames))
		for _, roleName := range roleNames {
			_, enabled := rolePermSets[roleName][entry.Key]
			roleValues[roleName] = enabled
		}
		rows = append(rows, &hubpermissionv1.PermissionRow{
			Key:        entry.Key,
			Category:   entry.Category,
			Label:      entry.Label,
			RoleValues: roleValues,
		})
	}

	return &hubpermissionv1.PermissionMatrix{
		Rows:  rows,
		Roles: roleNames,
	}
}

func toProtoRole(role domain.Role) *hubpermissionv1.CustomRole {
	perms := role.Permissions
	if perms == nil {
		perms = []string{}
	}
	return &hubpermissionv1.CustomRole{
		Id:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		Permissions:  perms,
		AppId:        role.AppID,
		RoleKey:      role.RoleKey,
		Level:        role.Level,
		ParentRoleId: role.ParentRoleID,
		ScopeRules:   role.ScopeRules,
		IsSystem:     role.IsSystem,
		IsArchived:   role.IsArchived,
		Lifecycle:    role.Lifecycle,
		Version:      role.Version,
		MemberCount:  role.MemberCount,
		CreatedAt:    role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toProtoAssignment(assignment domain.RoleAssignment) *hubpermissionv1.RoleAssignment {
	return &hubpermissionv1.RoleAssignment{
		Id:        assignment.ID,
		RoleId:    assignment.RoleID,
		AppId:     assignment.AppID,
		UserId:    assignment.UserID,
		GroupId:   assignment.GroupID,
		Status:    assignment.Status,
		Source:    assignment.Source,
		IsPrimary: assignment.IsPrimary,
		Scopes:    assignment.Scopes,
		GrantedAt: assignment.GrantedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toProtoAuditEvent(event domain.AuditEvent) *hubpermissionv1.RoleAuditEvent {
	return &hubpermissionv1.RoleAuditEvent{
		Id:            event.ID,
		Action:        event.Action,
		AppId:         event.AppID,
		RoleId:        event.RoleID,
		ActorUserId:   event.ActorUserID,
		TargetUserId:  event.TargetUserID,
		TargetGroupId: event.TargetGroupID,
		BeforeJson:    event.BeforeJSON,
		AfterJson:     event.AfterJSON,
		Reason:        event.Reason,
		CreatedAt:     event.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
