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
			if app.AppID == appID && app.HasAccess {
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
