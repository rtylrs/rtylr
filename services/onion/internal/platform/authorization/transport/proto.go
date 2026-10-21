package transport

import (
	"sort"
	"strconv"

	"github.com/voxire/onion/internal/platform/authorization"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

func EffectiveAccessToProto(access *authorization.EffectiveAccess) *hubuserv1.EffectiveAccess {
	if access == nil {
		return nil
	}

	result := &hubuserv1.EffectiveAccess{
		BusinessId:           strconv.FormatInt(access.BusinessID, 10),
		DashPermissions:      append([]string{}, access.DashPermissions...),
		SubscribedApps:       make([]string, 0, len(access.SubscribedApps)),
		Licenses:             make([]*hubuserv1.LicenseAssignment, 0, len(access.Licenses)),
		Apps:                 make([]*hubuserv1.EffectiveAppAccess, 0, len(access.Apps)),
		IsOwner:              access.IsOwner,
		IsAdmin:              access.IsAdmin,
		AuthorizationVersion: access.AuthorizationVersion,
	}
	for _, appID := range access.SubscribedApps {
		result.SubscribedApps = append(result.SubscribedApps, string(appID))
	}
	for _, license := range access.Licenses {
		status := "revoked"
		if license.Active {
			status = "active"
		}
		item := &hubuserv1.LicenseAssignment{
			AppId:            string(license.AppID),
			Status:           status,
			AssignmentSource: string(license.Source),
			GrantedAt:        license.GrantedAt,
			RevokedAt:        license.RevokedAt,
		}
		if license.GroupID != nil {
			item.GroupId = stringPointer(strconv.FormatInt(*license.GroupID, 10))
		}
		result.Licenses = append(result.Licenses, item)
	}
	for _, app := range access.Apps {
		item := &hubuserv1.EffectiveAppAccess{
			AppId:            string(app.AppID),
			Licensed:         app.Licensed,
			HasAccess:        app.HasAccess,
			Permissions:      append([]string{}, app.Permissions...),
			Sources:          make([]string, 0, len(app.Sources)),
			PermissionScopes: make([]*hubuserv1.PermissionScopeGrant, 0, len(app.PermissionScopes)),
		}
		for _, source := range app.Sources {
			item.Sources = append(item.Sources, string(source))
		}
		permissionKeys := make([]string, 0, len(app.PermissionScopes))
		for permission := range app.PermissionScopes {
			permissionKeys = append(permissionKeys, permission)
		}
		sort.Strings(permissionKeys)
		for _, permission := range permissionKeys {
			scopes := app.PermissionScopes[permission]
			grant := &hubuserv1.PermissionScopeGrant{PermissionKey: permission, Scopes: make([]*hubuserv1.PermissionScope, 0, len(scopes))}
			for _, scope := range scopes {
				grant.Scopes = append(grant.Scopes, &hubuserv1.PermissionScope{ScopeType: scope.Type, ScopeId: scope.ID})
			}
			item.PermissionScopes = append(item.PermissionScopes, grant)
		}
		result.Apps = append(result.Apps, item)
	}
	return result
}

func PermissionProfileAssignmentsToProto(assignments []authorization.RoleAssignment) []*hubuserv1.PermissionProfileAssignment {
	result := make([]*hubuserv1.PermissionProfileAssignment, 0, len(assignments))
	for _, assignment := range assignments {
		status := "revoked"
		if assignment.Active {
			status = "active"
		}
		item := &hubuserv1.PermissionProfileAssignment{
			AppId:                string(assignment.AppID),
			PermissionProfileIds: make([]string, 0, len(assignment.RoleIDs)),
			AssignmentSource:     string(assignment.Source),
			Status:               status,
			Scopes:               make([]*hubuserv1.PermissionScope, 0, len(assignment.Scopes)),
		}
		for _, roleID := range assignment.RoleIDs {
			item.PermissionProfileIds = append(item.PermissionProfileIds, strconv.FormatInt(roleID, 10))
		}
		if assignment.GroupID != nil {
			item.GroupId = stringPointer(strconv.FormatInt(*assignment.GroupID, 10))
		}
		for _, scope := range assignment.Scopes {
			item.Scopes = append(item.Scopes, &hubuserv1.PermissionScope{ScopeType: scope.Type, ScopeId: scope.ID})
		}
		result = append(result, item)
	}
	return result
}

func ProvisioningStatusToProto(status *authorization.ProvisioningStatus) *hubuserv1.ProvisioningStatus {
	if status == nil {
		return nil
	}
	return &hubuserv1.ProvisioningStatus{
		Mode:                   stringPointer(string(status.Mode)),
		State:                  stringPointer(string(status.State)),
		EmailVerified:          boolPointer(status.EmailVerified),
		PasswordChangeRequired: boolPointer(status.PasswordChangeRequired),
		MfaRequired:            boolPointer(status.MFARequired),
		MfaEnrolled:            boolPointer(status.MFAEnrolled),
		MfaVerified:            boolPointer(status.MFAVerified),
		RecoveryAvailable:      boolPointer(status.RecoveryAvailable),
	}
}

func stringPointer(value string) *string { return &value }

func boolPointer(value bool) *bool { return &value }
