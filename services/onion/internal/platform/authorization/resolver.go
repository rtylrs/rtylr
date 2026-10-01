// Package authorization contains the tenant authorization model and pure
// effective-access resolver. Persistence adapters translate database rows into
// Snapshot values before calling Resolve.
package authorization

import "sort"

type AppID string

const (
	AppHub      AppID = "hub"
	AppPOS      AppID = "pos"
	AppERP      AppID = "erp"
	AppHR       AppID = "hr"
	AppCRM      AppID = "crm"
	AppFinance  AppID = "finance"
	AppFlow     AppID = "flow"
	AppInsights AppID = "insights"
	AppMenu     AppID = "menu"
	AppRecruit  AppID = "recruit"
)

type AssignmentSource string

const (
	SourceExplicit AssignmentSource = "explicit"
	SourceGroup    AssignmentSource = "group"
	SourceSystem   AssignmentSource = "system"
)

type Scope struct {
	Type string
	ID   string
}

type Role struct {
	ID             int64
	AppID          *AppID
	Key            string
	DisplayName    string
	Description    string
	ParentRoleID   *int64
	Permissions    []string
	Level          int
	IsArchived     bool
	IsSystem       bool
	AssignmentMode string
}

type License struct {
	AppID     AppID
	Source    AssignmentSource
	Active    bool
	GroupID   *int64
	GrantedAt string
	RevokedAt *string
}

type RoleAssignment struct {
	AppID         AppID
	GroupID       *int64
	RoleIDs       []int64
	PrimaryRoleID *int64
	Source        AssignmentSource
	Scopes        []Scope
	Active        bool
}

type Snapshot struct {
	BusinessID           int64
	IsOwner              bool
	IsAdmin              bool
	MembershipActive     bool
	SubscribedApps       []AppID
	DashPermissions      []string
	Licenses             []License
	Assignments          []RoleAssignment
	Roles                []Role
	AllPermissions       []string
	AuthorizationVersion int64
}

type EffectiveAppAccess struct {
	AppID            AppID
	Licensed         bool
	HasAccess        bool
	Permissions      []string
	PermissionScopes map[string][]Scope
	Sources          []AssignmentSource
}

type EffectiveAccess struct {
	BusinessID           int64
	DashPermissions      []string
	SubscribedApps       []AppID
	Licenses             []License
	Apps                 []EffectiveAppAccess
	IsOwner              bool
	IsAdmin              bool
	AuthorizationVersion int64
}

type ProvisioningMode string

const (
	ProvisioningInvitation ProvisioningMode = "invitation"
	// #nosec G101 -- This label selects a provisioning mode; it is not a credential.
	ProvisioningDirectCredentials ProvisioningMode = "direct_credentials"
)

type ProvisioningState string

const (
	ProvisioningPendingInvitation ProvisioningState = "pending_invitation"
	ProvisioningPendingActivation ProvisioningState = "pending_activation"
	ProvisioningActive            ProvisioningState = "active"
	ProvisioningSuspended         ProvisioningState = "suspended"
	ProvisioningRevoked           ProvisioningState = "revoked"
)

type ProvisioningStatus struct {
	Mode                   ProvisioningMode
	State                  ProvisioningState
	EmailVerified          bool
	PasswordChangeRequired bool
	MFARequired            bool
	MFAEnrolled            bool
	MFAVerified            bool
	RecoveryAvailable      bool
}

// Resolve computes effective access from active membership, app entitlements,
// permission-profile assignments, and their resource scopes.
func Resolve(snapshot Snapshot) EffectiveAccess {
	isOwner := snapshot.MembershipActive && snapshot.IsOwner
	isAdmin := snapshot.MembershipActive && snapshot.IsAdmin
	rolesByID := make(map[int64]Role, len(snapshot.Roles))
	for _, role := range snapshot.Roles {
		rolesByID[role.ID] = role
	}

	dashPermissions := []string{}
	if snapshot.MembershipActive {
		// Every active business member can enter the Dash app launcher. Page
		// management remains permission-gated below this baseline.
		dashPermissions = uniqueSorted(append(snapshot.DashPermissions, "hub:dashboard:view"))
		if isOwner || isAdmin {
			dashPermissions = permissionsForApp(snapshot.AllPermissions, AppHub)
		}
	}

	apps := make([]AppID, 0, len(snapshot.SubscribedApps))
	seenApps := make(map[AppID]struct{}, len(snapshot.SubscribedApps))
	for _, appID := range snapshot.SubscribedApps {
		if _, seen := seenApps[appID]; seen {
			continue
		}
		seenApps[appID] = struct{}{}
		apps = append(apps, appID)
	}
	sort.Slice(apps, func(i, j int) bool { return apps[i] < apps[j] })

	resolved := EffectiveAccess{
		BusinessID:           snapshot.BusinessID,
		DashPermissions:      dashPermissions,
		SubscribedApps:       apps,
		Licenses:             append([]License{}, snapshot.Licenses...),
		Apps:                 make([]EffectiveAppAccess, 0, len(apps)),
		IsOwner:              isOwner,
		IsAdmin:              isAdmin,
		AuthorizationVersion: snapshot.AuthorizationVersion,
	}

	for _, appID := range apps {
		resolved.Apps = append(resolved.Apps, resolveApp(snapshot, rolesByID, appID, isOwner || isAdmin))
	}
	return resolved
}

// resolveApp combines the member's license and permission profile grants.
// Owners and admins inherit every permission for apps in the business plan.
func resolveApp(snapshot Snapshot, rolesByID map[int64]Role, appID AppID, isPrivileged bool) EffectiveAppAccess {
	licensed := false
	sources := make([]AssignmentSource, 0, 2)
	for _, license := range snapshot.Licenses {
		if license.AppID != appID || !license.Active || !validSourceTarget(license.Source, license.GroupID) {
			continue
		}
		licensed = true
		sources = appendUniqueSource(sources, license.Source)
	}
	// An owner or admin is entitled to the whole business plan, and their app
	// set is drawn from that plan, so every app they resolve is covered by it.
	// Without this the reported metadata contradicts the entitlement: the app
	// carries HasAccess but Licensed=false, and consumers that require both
	// (the Dash sidebar) then hide it.
	//
	// This is reporting only and cannot widen access: the caller's app set is
	// already bounded to the plan, and an inactive membership returns below
	// before either flag is set.
	if snapshot.MembershipActive && isPrivileged {
		licensed = true
		sources = appendUniqueSource(sources, SourceSystem)
	}

	access := EffectiveAppAccess{
		AppID:            appID,
		Licensed:         licensed,
		HasAccess:        false,
		Sources:          sources,
		Permissions:      []string{},
		PermissionScopes: map[string][]Scope{},
	}
	if !snapshot.MembershipActive {
		return access
	}

	if isPrivileged {
		access.HasAccess = true
		access.Permissions = permissionsForApp(snapshot.AllPermissions, appID)
		for _, permission := range access.Permissions {
			access.PermissionScopes[permission] = []Scope{}
		}
		access.Sources = appendUniqueSource(access.Sources, SourceSystem)
		return access
	}
	if !licensed {
		return access
	}

	for _, assignment := range snapshot.Assignments {
		if assignment.AppID != appID || !assignment.Active || !validSourceTarget(assignment.Source, assignment.GroupID) {
			continue
		}
		assignmentPermissions := make([]string, 0)
		validAssignment := false
		for _, roleID := range assignment.RoleIDs {
			role, ok := rolesByID[roleID]
			if !ok || role.IsArchived || role.AppID == nil || *role.AppID != appID {
				continue
			}
			validAssignment = true
			for _, permission := range inheritedPermissions(role, rolesByID, map[int64]struct{}{}) {
				assignmentPermissions = appendUniqueString(assignmentPermissions, permission)
			}
		}
		if !validAssignment {
			continue
		}
		access.HasAccess = true
		access.Sources = appendUniqueSource(access.Sources, assignment.Source)
		for _, permission := range assignmentPermissions {
			access.Permissions = appendUniqueString(access.Permissions, permission)
			if _, exists := access.PermissionScopes[permission]; !exists {
				access.PermissionScopes[permission] = append([]Scope{}, assignment.Scopes...)
				continue
			}
			current := access.PermissionScopes[permission]
			if len(current) == 0 || len(assignment.Scopes) == 0 {
				access.PermissionScopes[permission] = []Scope{}
				continue
			}
			for _, scope := range assignment.Scopes {
				access.PermissionScopes[permission] = appendUniqueScope(access.PermissionScopes[permission], scope)
			}
		}
	}
	access.Permissions = uniqueSorted(access.Permissions)
	return access
}

func validSourceTarget(source AssignmentSource, groupID *int64) bool {
	switch source {
	case SourceExplicit, SourceSystem:
		return groupID == nil
	case SourceGroup:
		return groupID != nil && *groupID > 0
	default:
		return false
	}
}

func permissionsForApp(permissions []string, appID AppID) []string {
	prefix := string(appID) + ":"
	if appID == AppHub {
		prefix = "hub:"
	}
	filtered := make([]string, 0)
	for _, permission := range permissions {
		if len(permission) >= len(prefix) && permission[:len(prefix)] == prefix {
			filtered = append(filtered, permission)
		}
	}
	return uniqueSorted(filtered)
}

// PermissionsForRole returns the inherited permission set for a role. It is
// exported for persistence adapters that hydrate a snapshot from normalized
// role and permission rows before resolving the complete access payload.
func PermissionsForRole(roleID int64, roles []Role) []string {
	rolesByID := make(map[int64]Role, len(roles))
	for _, role := range roles {
		rolesByID[role.ID] = role
	}
	role, ok := rolesByID[roleID]
	if !ok || role.IsArchived {
		return []string{}
	}
	return inheritedPermissions(role, rolesByID, map[int64]struct{}{})
}

func inheritedPermissions(role Role, rolesByID map[int64]Role, seen map[int64]struct{}) []string {
	if _, ok := seen[role.ID]; ok {
		return nil
	}
	seen[role.ID] = struct{}{}
	permissions := append([]string{}, role.Permissions...)
	if role.ParentRoleID == nil {
		return uniqueSorted(permissions)
	}
	parent, ok := rolesByID[*role.ParentRoleID]
	if !ok || parent.IsArchived {
		return uniqueSorted(permissions)
	}
	return uniqueSorted(append(permissions, inheritedPermissions(parent, rolesByID, seen)...))
}

func appendUniqueSource(values []AssignmentSource, value AssignmentSource) []AssignmentSource {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueInt64(values []int64, value int64) []int64 {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueScope(values []Scope, value Scope) []Scope {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func uniqueSorted(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = appendUniqueString(result, value)
	}
	sort.Strings(result)
	return result
}
