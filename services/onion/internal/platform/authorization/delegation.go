package authorization

// CanDelegatePermissions reports whether an actor may grant the supplied
// permissions for an application. Owners and admins are the delegation
// ceiling for the tenant. Other actors must already hold every permission
// they would grant, and must have active access to that application.
func CanDelegatePermissions(access *EffectiveAccess, appID AppID, permissions []string) bool {
	if access == nil {
		return false
	}
	if access.IsOwner || access.IsAdmin {
		return true
	}

	granted := access.DashPermissions
	if appID != "" && appID != "dash" {
		appFound := false
		granted = nil
		for _, app := range access.Apps {
			if app.AppID == appID && app.Licensed && app.HasAccess {
				appFound = true
				granted = app.Permissions
				break
			}
		}
		if !appFound {
			return false
		}
	}

	for _, permission := range permissions {
		if !containsPermission(granted, permission) {
			return false
		}
	}
	return true
}

func containsPermission(permissions []string, target string) bool {
	for _, permission := range permissions {
		if permission == target {
			return true
		}
	}
	return false
}

// DashboardRoleLevel returns the ordered ceiling for the built-in Dash roles.
// A larger value is more privileged. Custom role levels are checked against
// their permission set by the caller and are not represented here.
func DashboardRoleLevel(role string) int {
	switch role {
	case "owner":
		return 5
	case "admin":
		return 4
	case "manager":
		return 3
	case "member":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

// CanAssignDashboardRole prevents a non-admin actor from assigning a role at
// or above their own built-in Dash level. Owners and admins are handled by the
// general delegation ceiling.
func CanAssignDashboardRole(access *EffectiveAccess, role string) bool {
	if access == nil {
		return false
	}
	if access.IsOwner || access.IsAdmin {
		return true
	}
	actorLevel := DashboardRoleLevel(access.DashRole)
	targetLevel := DashboardRoleLevel(role)
	return actorLevel > 0 && targetLevel > 0 && targetLevel < actorLevel
}
