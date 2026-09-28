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
	SourceExplicit    AssignmentSource = "explicit"
	SourceDashDefault AssignmentSource = "dash_default"
	SourceGroup       AssignmentSource = "group"
	SourceSystem      AssignmentSource = "system"
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
	DashRole             string
	DashRoleID           *int64
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
	AppID             AppID
	Licensed          bool
	HasAccess         bool
	RoleIDs           []int64
	PrimaryRoleID     *int64
	DisplayRole       string
	Permissions       []string
	Scopes            []Scope
	Sources           []AssignmentSource
	PermissionSources map[string][]string
}

type EffectiveAccess struct {
	BusinessID           int64
	DashRole             string
	DashRoleID           *int64
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

// Resolve computes effective access using union semantics for compatible
// assignments. It deliberately returns no app permission when membership is
// inactive, a license is absent, a role is archived, or a role cannot be found.
func Resolve(snapshot Snapshot) EffectiveAccess {
	// A revoked business membership must not retain dashboard privileges merely
	// because the stale dash role/permission rows are still present. App access
	// already fails closed on MembershipActive; apply the same boundary to the
	// dashboard surface and its owner/admin shortcuts.
	isOwner := snapshot.MembershipActive && snapshot.DashRole == "owner"
	isAdmin := snapshot.MembershipActive && snapshot.DashRole == "admin"
	rolesByID := make(map[int64]Role, len(snapshot.Roles))
	for _, role := range snapshot.Roles {
		rolesByID[role.ID] = role
	}

	dashPermissions := []string{}
	if snapshot.MembershipActive {
		dashPermissions = uniqueSorted(snapshot.DashPermissions)
		if isOwner || isAdmin {
			dashPermissions = uniqueSorted(snapshot.AllPermissions)
		}
	}

	dashRole := snapshot.DashRole
	dashRoleID := snapshot.DashRoleID
	if !snapshot.MembershipActive {
		// CanAssignDashboardRole and other consumers use the role string as a
		// fallback ceiling, so clear the identity as well as the flags.
		dashRole = ""
		dashRoleID = nil
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
		DashRole:             dashRole,
		DashRoleID:           dashRoleID,
		DashPermissions:      dashPermissions,
		SubscribedApps:       apps,
		Licenses:             append([]License{}, snapshot.Licenses...),
		Apps:                 make([]EffectiveAppAccess, 0, len(apps)),
		IsOwner:              isOwner,
		IsAdmin:              isAdmin,
		AuthorizationVersion: snapshot.AuthorizationVersion,
	}

	for _, appID := range apps {
		resolved.Apps = append(resolved.Apps, resolveApp(snapshot, rolesByID, appID, isOwner, isAdmin))
	}
	return resolved
}

// LegacyAccess reconstructs the pre-v2 app-level boundary for rollback. The
// old model had no feature-level permission set, so an active legacy app grant
// authorizes the app's existing surface. This is intentionally kept separate
// from Resolve so the migration can be reversed without deleting normalized
// assignments, licenses, or audit history.
func LegacyAccess(businessID int64, apps []string, role string) EffectiveAccess {
	knownApps := map[AppID]struct{}{
		AppPOS: {}, AppERP: {}, AppHR: {}, AppCRM: {}, AppFinance: {},
		AppFlow: {}, AppInsights: {}, AppMenu: {}, AppRecruit: {},
	}
	resolvedApps := make([]AppID, 0, len(apps))
	licenses := make([]License, 0, len(apps))
	seen := make(map[AppID]struct{}, len(apps))
	for _, rawApp := range apps {
		appID := AppID(rawApp)
		if _, ok := knownApps[appID]; !ok {
			continue
		}
		if _, ok := seen[appID]; ok {
			continue
		}
		seen[appID] = struct{}{}
		resolvedApps = append(resolvedApps, appID)
		licenses = append(licenses, License{AppID: appID, Source: SourceExplicit, Active: true})
	}
	sort.Slice(resolvedApps, func(i, j int) bool { return resolvedApps[i] < resolvedApps[j] })
	access := EffectiveAccess{
		BusinessID:     businessID,
		DashRole:       role,
		SubscribedApps: append([]AppID{}, resolvedApps...),
		Licenses:       licenses,
		Apps:           make([]EffectiveAppAccess, 0, len(resolvedApps)),
		IsOwner:        role == "owner",
		IsAdmin:        role == "admin",
	}
	for _, appID := range resolvedApps {
		access.Apps = append(access.Apps, EffectiveAppAccess{
			AppID:             appID,
			Licensed:          true,
			HasAccess:         true,
			Sources:           []AssignmentSource{SourceExplicit},
			Permissions:       []string{},
			Scopes:            []Scope{},
			PermissionSources: map[string][]string{},
		})
	}
	return access
}

// resolveApp reports one app's effective access. Role assignments are the
// authorization boundary: a license row is plan metadata and never grants or
// denies access, because a self-hosted install can legitimately have no license
// rows at all. The owner/admin system path grants the full catalog for every
// app the caller is already entitled to; which apps those are is decided by
// the caller's app set, not by licensing.
func resolveApp(snapshot Snapshot, rolesByID map[int64]Role, appID AppID, isOwner, isAdmin bool) EffectiveAppAccess {
	licensed := false
	sources := make([]AssignmentSource, 0, 2)
	for _, license := range snapshot.Licenses {
		if license.AppID != appID || !license.Active || !validSourceTarget(license.Source, license.GroupID) {
			continue
		}
		licensed = true
		sources = appendUniqueSource(sources, license.Source)
	}

	access := EffectiveAppAccess{
		AppID:             appID,
		Licensed:          licensed,
		HasAccess:         false,
		RoleIDs:           []int64{},
		Scopes:            []Scope{},
		Sources:           sources,
		Permissions:       []string{},
		PermissionSources: map[string][]string{},
	}
	if !snapshot.MembershipActive {
		return access
	}

	if isOwner || isAdmin {
		access.HasAccess = true
		access.Permissions = uniqueSorted(snapshot.AllPermissions)
		access.Sources = appendUniqueSource(access.Sources, SourceSystem)
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
			access.RoleIDs = appendUniqueInt64(access.RoleIDs, roleID)
			if assignment.PrimaryRoleID != nil && *assignment.PrimaryRoleID == roleID {
				access.PrimaryRoleID = assignment.PrimaryRoleID
				access.DisplayRole = role.DisplayName
			}
			for _, permission := range inheritedPermissions(role, rolesByID, map[int64]struct{}{}) {
				assignmentPermissions = appendUniqueString(assignmentPermissions, permission)
				access.PermissionSources[permission] = appendUniqueString(
					access.PermissionSources[permission], role.Key,
				)
			}
		}
		if !validAssignment {
			continue
		}
		access.HasAccess = true
		access.Sources = appendUniqueSource(access.Sources, assignment.Source)
		for _, scope := range assignment.Scopes {
			access.Scopes = appendUniqueScope(access.Scopes, scope)
		}
		for _, permission := range assignmentPermissions {
			access.Permissions = appendUniqueString(access.Permissions, permission)
		}
	}
	access.RoleIDs = sortedInt64(access.RoleIDs)
	access.Permissions = uniqueSorted(access.Permissions)
	return access
}

func validSourceTarget(source AssignmentSource, groupID *int64) bool {
	switch source {
	case SourceExplicit, SourceDashDefault, SourceSystem:
		return groupID == nil
	case SourceGroup:
		return groupID != nil && *groupID > 0
	default:
		return false
	}
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

func sortedInt64(values []int64) []int64 {
	result := append([]int64{}, values...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
