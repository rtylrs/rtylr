package authorization_test

import (
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

type matrixPermission struct {
	app        authorization.AppID
	permission string
}

var featureMatrixPermissions = []matrixPermission{
	{authorization.AppHub, "hub:dashboard:view"},
	{authorization.AppHub, "hub:team:view"},
	{authorization.AppHub, "hub:team:invite"},
	{authorization.AppHub, "hub:team:edit"},
	{authorization.AppHub, "hub:team:deactivate"},
	{authorization.AppHub, "hub:roles:view"},
	{authorization.AppHub, "hub:roles:create"},
	{authorization.AppHub, "hub:roles:edit"},
	{authorization.AppHub, "hub:roles:archive"},
	{authorization.AppHub, "hub:licenses:manage"},
	{authorization.AppHub, "hub:groups:manage"},
	{authorization.AppHub, "hub:hierarchy:manage"},
	{authorization.AppHub, "hub:audit:view"},
	{authorization.AppHub, "hub:security:manage"},
	{authorization.AppHub, "hub:api_keys:manage"},
	{authorization.AppPOS, "pos:orders:view"},
	{authorization.AppPOS, "pos:orders:create"},
	{authorization.AppPOS, "pos:orders:void"},
	{authorization.AppPOS, "pos:orders:void:approve"},
	{authorization.AppPOS, "pos:orders:refund"},
	{authorization.AppPOS, "pos:orders:refund:approve"},
	{authorization.AppPOS, "pos:orders:void:override"},
	{authorization.AppPOS, "pos:orders:refund:override"},
	{authorization.AppPOS, "pos:orders:discount"},
	{authorization.AppPOS, "pos:cash:view"},
	{authorization.AppPOS, "pos:cash:drop"},
	{authorization.AppPOS, "pos:cash:pickup"},
	{authorization.AppPOS, "pos:cash:approve"},
	{authorization.AppPOS, "pos:employees:manage"},
	{authorization.AppPOS, "pos:terminals:manage"},
	{authorization.AppPOS, "pos:reports:view"},
	{authorization.AppERP, "erp:products:view"},
	{authorization.AppERP, "erp:products:create"},
	{authorization.AppERP, "erp:products:edit"},
	{authorization.AppERP, "erp:products:delete"},
	{authorization.AppERP, "erp:inventory:view"},
	{authorization.AppERP, "erp:inventory:adjust"},
	{authorization.AppERP, "erp:inventory:transfer"},
	{authorization.AppERP, "erp:inventory:approve"},
	{authorization.AppERP, "erp:warehouses:manage"},
	{authorization.AppERP, "erp:purchasing:view"},
	{authorization.AppERP, "erp:purchasing:create"},
	{authorization.AppERP, "erp:purchasing:approve"},
	{authorization.AppFinance, "finance:invoices:view"},
	{authorization.AppFinance, "finance:invoices:create"},
	{authorization.AppFinance, "finance:invoices:send"},
	{authorization.AppFinance, "finance:expenses:view"},
	{authorization.AppFinance, "finance:expenses:create"},
	{authorization.AppFinance, "finance:expenses:approve"},
	{authorization.AppFinance, "finance:bank_feeds:view"},
	{authorization.AppFinance, "finance:bank_feeds:sensitive_view"},
	{authorization.AppFinance, "finance:reconciliation:manage"},
	{authorization.AppFinance, "finance:reports:view"},
	{authorization.AppFinance, "finance:reports:export"},
	{authorization.AppFinance, "finance:tax:configure"},
	{authorization.AppCRM, "crm:customers:view"},
	{authorization.AppCRM, "crm:customers:create"},
	{authorization.AppCRM, "crm:customers:edit"},
	{authorization.AppCRM, "crm:customers:delete"},
	{authorization.AppCRM, "crm:customers:export"},
	{authorization.AppCRM, "crm:campaigns:view"},
	{authorization.AppCRM, "crm:campaigns:create"},
	{authorization.AppCRM, "crm:campaigns:send"},
	{authorization.AppCRM, "crm:loyalty:manage"},
	{authorization.AppCRM, "crm:automations:manage"},
	{authorization.AppHR, "hr:employees:view"},
	{authorization.AppHR, "hr:employees:manage"},
	{authorization.AppHR, "hr:employees:sensitive_view"},
	{authorization.AppHR, "hr:time_off:request"},
	{authorization.AppHR, "hr:time_off:approve"},
	{authorization.AppHR, "hr:schedules:manage"},
	{authorization.AppHR, "hr:payroll:view"},
	{authorization.AppHR, "hr:payroll:manage"},
	{authorization.AppHR, "hr:payroll:execute"},
	{authorization.AppHR, "hr:roles:manage"},
	{authorization.AppFlow, "flow:projects:view"},
	{authorization.AppFlow, "flow:projects:create"},
	{authorization.AppFlow, "flow:projects:edit"},
	{authorization.AppFlow, "flow:projects:assign"},
	{authorization.AppFlow, "flow:tasks:view"},
	{authorization.AppFlow, "flow:tasks:create"},
	{authorization.AppFlow, "flow:tasks:edit"},
	{authorization.AppFlow, "flow:tasks:assign"},
	{authorization.AppFlow, "flow:team:manage"},
	{authorization.AppFlow, "flow:managers:manage"},
	{authorization.AppFlow, "flow:settings:manage"},
	{authorization.AppInsights, "insights:dashboard:view"},
	{authorization.AppInsights, "insights:financial:view"},
	{authorization.AppInsights, "insights:customers:view"},
	{authorization.AppInsights, "insights:reports:create"},
	{authorization.AppInsights, "insights:reports:share"},
	{authorization.AppInsights, "insights:reports:export"},
	{authorization.AppMenu, "menu:menus:view"},
	{authorization.AppMenu, "menu:menus:edit"},
	{authorization.AppMenu, "menu:branding:edit"},
	{authorization.AppMenu, "menu:orders:view"},
	{authorization.AppMenu, "menu:analytics:view"},
	{authorization.AppRecruit, "recruit:jobs:view"},
	{authorization.AppRecruit, "recruit:jobs:manage"},
	{authorization.AppRecruit, "recruit:jobs:publish"},
	{authorization.AppRecruit, "recruit:candidates:view"},
	{authorization.AppRecruit, "recruit:candidates:edit"},
	{authorization.AppRecruit, "recruit:candidates:sensitive_view"},
	{authorization.AppRecruit, "recruit:candidates:export"},
}

func TestFeatureMatrixContainsExactly104UniquePermissions(t *testing.T) {
	if len(featureMatrixPermissions) != 104 {
		t.Fatalf("feature matrix permissions = %d, want 104", len(featureMatrixPermissions))
	}
	seen := make(map[string]struct{}, len(featureMatrixPermissions))
	for _, entry := range featureMatrixPermissions {
		if _, exists := seen[entry.permission]; exists {
			t.Fatalf("duplicate permission %q", entry.permission)
		}
		seen[entry.permission] = struct{}{}
	}
}

func TestFeatureMatrixPermissionsEnforceAppLicenseAndExactPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	for _, entry := range featureMatrixPermissions {
		entry := entry
		t.Run(entry.permission, func(t *testing.T) {
			active := featureMatrixUser(entry, true, true, []string{entry.permission})
			if !authPkg.HasAppPermission(active, entry.app, entry.permission) {
				t.Fatal("minimum licensed assignment was denied")
			}

			for name, user := range map[string]*authPkg.AuthUser{
				"missing license":    featureMatrixUser(entry, false, false, []string{entry.permission}),
				"missing permission": featureMatrixUser(entry, true, true, nil),
				"wrong app":          featureMatrixUser(matrixPermission{app: differentApp(entry.app), permission: entry.permission}, true, true, []string{entry.permission}),
				"stale session":      &authPkg.AuthUser{UserID: 7, BusinessID: 9},
			} {
				if authPkg.HasAppPermission(user, entry.app, entry.permission) {
					t.Errorf("%s unexpectedly retained %s", name, entry.permission)
				}
			}

			inactive := featureMatrixResolvedUser(entry, false, false)
			if authPkg.HasAppPermission(inactive, entry.app, entry.permission) {
				t.Fatal("inactive membership retained application permission")
			}
			archived := featureMatrixResolvedUser(entry, true, true)
			if authPkg.HasAppPermission(archived, entry.app, entry.permission) {
				t.Fatal("archived role retained application permission")
			}
		})
	}
}

func TestFeatureMatrixPermissionsFailClosedForLifecycleAndGroupStates(t *testing.T) {
	for _, entry := range featureMatrixPermissions {
		entry := entry
		t.Run(entry.permission, func(t *testing.T) {
			cases := map[string]authorization.Snapshot{
				"revoked assignment":   featureMatrixSnapshot(entry, true, false, true, true, false, false),
				"removed license":      featureMatrixSnapshot(entry, false, true, true, true, false, false),
				"deactivated user":     featureMatrixSnapshot(entry, true, true, false, true, false, false),
				"terminated employee":  featureMatrixSnapshot(entry, true, true, false, true, false, false),
				"invalid group source": featureMatrixSnapshot(entry, true, true, true, true, true, false),
			}
			for name, snapshot := range cases {
				access := authorization.Resolve(snapshot)
				user := &authPkg.AuthUser{UserID: 7, BusinessID: 9, EffectiveAccess: &access}
				if authPkg.HasAppPermission(user, entry.app, entry.permission) {
					t.Errorf("%s retained %s", name, entry.permission)
				}
			}
		})
	}
}

func TestFeatureMatrixDelegationCeiling(t *testing.T) {
	for _, entry := range featureMatrixPermissions {
		entry := entry
		t.Run(entry.permission, func(t *testing.T) {
			access := featureMatrixUser(entry, true, true, []string{entry.permission}).EffectiveAccess
			if !authorization.CanDelegatePermissions(access, entry.app, []string{entry.permission}) {
				t.Fatal("an actor must be able to delegate a permission they hold")
			}
			if authorization.CanDelegatePermissions(access, entry.app, []string{"permission:not-held"}) {
				t.Fatal("an actor must not delegate a permission they do not hold")
			}
			if authorization.CanDelegatePermissions(&authorization.EffectiveAccess{}, entry.app, []string{entry.permission}) {
				t.Fatal("an unlicensed actor must not delegate an application permission")
			}
		})
	}

	for name, access := range map[string]*authorization.EffectiveAccess{
		"owner": {IsOwner: true},
		"admin": {IsAdmin: true},
	} {
		if !authorization.CanDelegatePermissions(access, authorization.AppHub, []string{"hub:roles:edit"}) {
			t.Errorf("%s must be able to delegate below the tenant delegation ceiling", name)
		}
	}
	if authorization.CanDelegatePermissions(nil, authorization.AppHub, []string{"hub:roles:edit"}) {
		t.Fatal("nil access must not delegate permissions")
	}
}

func featureMatrixUser(entry matrixPermission, licensed, hasAccess bool, permissions []string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       entry.app,
				Licensed:    licensed,
				HasAccess:   hasAccess,
				Permissions: permissions,
			}},
		},
	}
}

func differentApp(app authorization.AppID) authorization.AppID {
	if app == authorization.AppHub {
		return authorization.AppPOS
	}
	return authorization.AppHub
}

func featureMatrixResolvedUser(entry matrixPermission, membershipActive, archived bool) *authPkg.AuthUser {
	access := authorization.Resolve(featureMatrixSnapshot(entry, membershipActive, membershipActive, membershipActive, true, false, archived))
	return &authPkg.AuthUser{UserID: 7, BusinessID: 9, EffectiveAccess: &access}
}

func featureMatrixSnapshot(entry matrixPermission, licensed, assignmentActive, membershipActive, roleValid, invalidSource, archived bool) authorization.Snapshot {
	appID := entry.app
	roleAppID := &appID
	if !roleValid {
		wrongApp := differentApp(entry.app)
		roleAppID = &wrongApp
	}
	groupID := int64(100)
	source := authorization.SourceExplicit
	if invalidSource {
		source = authorization.AssignmentSource("invalid-group-source")
	}
	var licenseGroupID *int64
	var assignmentGroupID *int64
	if source == authorization.SourceGroup {
		licenseGroupID = &groupID
		assignmentGroupID = &groupID
	}
	return authorization.Snapshot{
		BusinessID:       9,
		MembershipActive: membershipActive,
		SubscribedApps:   []authorization.AppID{entry.app},
		Licenses:         []authorization.License{{AppID: entry.app, Active: licensed, Source: source, GroupID: licenseGroupID}},
		Assignments:      []authorization.RoleAssignment{{AppID: entry.app, GroupID: assignmentGroupID, RoleIDs: []int64{1}, Source: source, Active: assignmentActive}},
		Roles:            []authorization.Role{{ID: 1, AppID: roleAppID, Key: "matrix_role", Permissions: []string{entry.permission}, IsArchived: archived}},
	}
}
