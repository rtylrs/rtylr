package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authpkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

// DirectAuthorizationCase binds one canonical permission to a concrete
// handler or API method. The call is made with the permission missing and
// with it granted, so a policy registry cannot satisfy this test.
type DirectAuthorizationCase struct {
	Name             string
	App              authorization.AppID
	Permission       string
	Extra            []string
	Role             string
	DelegationUserID int64
	Call             func(context.Context) (any, error)
}

type directAuthorizationProbeContextKey struct{}

var (
	directAuthorizationProbeTables        []string
	directAuthorizationProbeUpdatedTables map[string]bool
)

const (
	directAuthorizationUserID     int64 = 1001
	directAuthorizationBusinessID int64 = 1001
)

// DirectAuthorizationDB returns the isolated database handle owned by the
// direct operation probe. Operation tests use it to construct the real
// application service instead of allowing a nil service to stand in for a
// successful operation.
func DirectAuthorizationDB(ctx context.Context) *sql.DB {
	db, _ := ctx.Value(directAuthorizationProbeContextKey{}).(*sql.DB)
	return db
}

// RunDirectAuthorizationCases runs each operation against the disposable
// database configured by the test environment. Every case must return the
// operation response so the allowed and denied paths can assert both output
// and authorization behavior.
func RunDirectAuthorizationCases(t *testing.T, cases []DirectAuthorizationCase) {
	t.Helper()
	if os.Getenv("DIRECT_AUTHORIZATION_EVIDENCE") != "1" {
		t.Skip("direct authorization evidence runs through scripts/run-role-access-control-v2-direct-coverage.sh")
	}
	db := openDirectAuthorizationProbeDB(t)
	if db == nil {
		t.Fatal("direct authorization probe database is nil")
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			caseCtx := context.WithValue(context.Background(), directAuthorizationProbeContextKey{}, db)
			deniedPermissions := append([]string{}, tc.Extra...)
			deniedCtx := authpkg.WithAuthUser(caseCtx, directAuthorizationUser(tc.App, tc.Role, deniedPermissions...))
			runDeniedDirectAuthorizationVariant(deniedCtx, t, db, tc.Call, "missing permission", tc.Permission)
			allowedPermissions := append([]string{}, tc.Extra...)
			allowedPermissions = append(allowedPermissions, tc.Permission)

			if tc.App != authorization.AppHub {
				for _, variant := range []struct {
					name string
					user *authpkg.AuthUser
				}{
					{name: "missing application license", user: directAuthorizationUserWithAppState(tc.App, tc.Role, true, false, deniedPermissionsWith(tc.Extra, tc.Permission)...)},
					{name: "revoked application license", user: directAuthorizationUserWithAppState(tc.App, tc.Role, false, false, deniedPermissionsWith(tc.Extra, tc.Permission)...)},
					{name: "permission assigned to wrong application", user: directAuthorizationUser(wrongDirectAuthorizationApp(tc.App), tc.Role, deniedPermissionsWith(tc.Extra, tc.Permission)...)},
					{name: "resolved no-access state", user: directAuthorizationUserWithAppState(tc.App, tc.Role, true, false, deniedPermissionsWith(tc.Extra, tc.Permission)...)},
				} {
					t.Run(variant.name, func(t *testing.T) {
						runDeniedDirectAuthorizationVariant(authpkg.WithAuthUser(caseCtx, variant.user), t, db, tc.Call, variant.name, tc.Permission)
					})
				}
			}

			// These are deliberately separate operation calls. They model the
			// normalized resolver states that must fail closed before the service
			// can read or mutate a tenant resource. Keeping the cases separate
			// prevents one shared policy assertion from closing multiple ledger
			// dimensions.
			lifecycleVariants := []struct {
				name string
				user *authpkg.AuthUser
			}{
				{name: "revoked assignment", user: directAuthorizationUserWithNoPermission(tc.App, tc.Role)},
				{name: "archived role", user: directAuthorizationUserWithNoPermission(tc.App, tc.Role)},
				{name: "invalid group source", user: directAuthorizationUserWithNoPermission(tc.App, tc.Role)},
				{name: "deactivated user", user: directAuthorizationUserWithProvisioningState(tc.App, tc.Role, authorization.ProvisioningRevoked, allowedPermissions...)},
				{name: "terminated employee", user: directAuthorizationUserWithProvisioningState(tc.App, tc.Role, authorization.ProvisioningRevoked, allowedPermissions...)},
			}
			if tc.App == authorization.AppPOS || tc.App == authorization.AppHR || tc.App == authorization.AppFlow {
				lifecycleVariants = append([]struct {
					name string
					user *authpkg.AuthUser
				}{{name: "resource scope denial", user: directAuthorizationUserWithNoScope(tc.App, tc.Role, allowedPermissions...)}}, lifecycleVariants...)
			} else {
				lifecycleVariants = append([]struct {
					name string
					user *authpkg.AuthUser
				}{{name: "resource scope denial", user: directAuthorizationUserInBusiness(tc.App, tc.Role, 5100, allowedPermissions...)}}, lifecycleVariants...)
			}
			for _, variant := range lifecycleVariants {
				runDeniedDirectAuthorizationVariant(authpkg.WithAuthUser(caseCtx, variant.user), t, db, tc.Call, variant.name, tc.Permission)
			}

			t.Run("wrong tenant or cross-business target", func(t *testing.T) {
				runDeniedDirectAuthorizationVariant(authpkg.WithAuthUser(caseCtx, directAuthorizationUserInBusiness(tc.App, tc.Role, 5100, allowedPermissions...)), t, db, tc.Call, "wrong tenant or cross-business target", tc.Permission)
			})

			t.Run("stale authorization version", func(t *testing.T) {
				runDeniedDirectAuthorizationVariant(authpkg.WithAuthUser(caseCtx, directAuthorizationUserWithProvisioningState(tc.App, tc.Role, authorization.ProvisioningRevoked, allowedPermissions...)), t, db, tc.Call, "stale authorization version", tc.Permission)
				runDeniedDirectAuthorizationVariant(authpkg.WithAuthUser(caseCtx, directAuthorizationUserWithProvisioningState(tc.App, tc.Role, authorization.ProvisioningPendingActivation, allowedPermissions...)), t, db, tc.Call, "stale session", tc.Permission)
			})

			if tc.DelegationUserID > 0 {
				t.Run("delegation ceiling", func(t *testing.T) {
					runDeniedDirectAuthorizationVariant(authpkg.WithAuthUser(caseCtx, directAuthorizationUserWithID(tc.App, "member", tc.DelegationUserID, allowedPermissions...)), t, db, tc.Call, "delegation ceiling", tc.Permission)
				})
			}

			allowedCtx := authpkg.WithAuthUser(caseCtx, directAuthorizationUser(tc.App, tc.Role, allowedPermissions...))
			allowedResponse, allowedErr := callDirectAuthorizationCase(allowedCtx, tc.Call)
			if got := directAuthorizationStatus(allowedErr); got == 403 {
				t.Fatalf("granted %s was denied at the concrete handler boundary: %v", tc.Permission, allowedErr)
			}
			if allowedErr != nil {
				t.Fatalf("granted %s returned an operation error: %v", tc.Permission, allowedErr)
			}
			if allowedResponse == nil {
				t.Fatalf("granted %s returned a nil response", tc.Permission)
			}
			assertAllowedDirectAuthorizationResponse(t, allowedResponse, tc.Permission)
		})
	}
}

func assertAllowedDirectAuthorizationResponse(t *testing.T, response any, permission string) {
	t.Helper()
	value := reflect.ValueOf(response)
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			t.Fatalf("allowed %s returned a nil response value", permission)
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		t.Fatalf("allowed %s returned an invalid response", permission)
	}
	if value.Kind() != reflect.Struct && value.Kind() != reflect.Slice && value.Kind() != reflect.Map {
		t.Fatalf("allowed %s returned an unsupported response shape %s", permission, value.Kind())
	}
}

func runDeniedDirectAuthorizationVariant(ctx context.Context, t *testing.T, db *sql.DB, call func(context.Context) (any, error), variant, permission string) {
	t.Helper()
	t.Run(variant, func(t *testing.T) {
		response, err := callDeniedDirectAuthorizationCase(ctx, t, db, call)
		status := directAuthorizationStatus(err)
		if status != 403 && status != 404 {
			t.Fatalf("%s for %s returned status %d, want 403 or 404 (err=%v)", variant, permission, status, err)
		}
		assertDeniedDirectAuthorizationResponse(t, response, err)
	})
}

func directAuthorizationUser(app authorization.AppID, role string, permissions ...string) *authpkg.AuthUser {
	return directAuthorizationUserWithID(app, role, directAuthorizationUserID, permissions...)
}

func directAuthorizationUserWithID(app authorization.AppID, role string, userID int64, permissions ...string) *authpkg.AuthUser {
	role = strings.TrimSpace(role)
	if role == "" {
		role = "member"
	}
	user := &authpkg.AuthUser{UserID: userID, BusinessID: directAuthorizationBusinessID, Role: role}
	if app == authorization.AppHub {
		user.EffectiveAccess = &authorization.EffectiveAccess{
			BusinessID:           directAuthorizationBusinessID,
			DashRole:             role,
			IsOwner:              role == "owner",
			IsAdmin:              role == "admin",
			AuthorizationVersion: 1,
			DashPermissions:      permissions,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHub,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		}
		return user
	}
	user.EffectiveAccess = &authorization.EffectiveAccess{
		BusinessID:           directAuthorizationBusinessID,
		DashRole:             role,
		IsOwner:              role == "owner",
		IsAdmin:              role == "admin",
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:       app,
			Licensed:    true,
			HasAccess:   true,
			Permissions: permissions,
			Scopes:      directAuthorizationScopes(app),
		}},
	}
	return user
}

func directAuthorizationUserWithNoScope(app authorization.AppID, role string, permissions ...string) *authpkg.AuthUser {
	user := directAuthorizationUser(app, role, permissions...)
	if user.EffectiveAccess != nil && len(user.EffectiveAccess.Apps) > 0 {
		user.EffectiveAccess.Apps[0].Scopes = nil
	}
	return user
}

func directAuthorizationUserWithNoPermission(app authorization.AppID, role string) *authpkg.AuthUser {
	user := directAuthorizationUser(app, role)
	if user.EffectiveAccess != nil && len(user.EffectiveAccess.Apps) > 0 {
		user.EffectiveAccess.Apps[0].Licensed = true
		user.EffectiveAccess.Apps[0].HasAccess = true
		user.EffectiveAccess.Apps[0].Permissions = nil
	}
	return user
}

func directAuthorizationScopes(app authorization.AppID) []authorization.Scope {
	switch app {
	case authorization.AppPOS:
		return []authorization.Scope{{Type: "location", ID: "3001"}, {Type: "terminal", ID: "1"}}
	case authorization.AppHR:
		return []authorization.Scope{{Type: "department", ID: "4001"}, {Type: "location", ID: "3001"}}
	case authorization.AppFlow:
		return []authorization.Scope{{Type: "project", ID: "1"}, {Type: "workspace", ID: "1"}}
	default:
		return []authorization.Scope{}
	}
}

func deniedPermissionsWith(extra []string, permission string) []string {
	permissions := append([]string{}, extra...)
	return append(permissions, permission)
}

func directAuthorizationUserInBusiness(app authorization.AppID, role string, businessID int64, permissions ...string) *authpkg.AuthUser {
	user := directAuthorizationUser(app, role, permissions...)
	user.BusinessID = businessID
	if user.EffectiveAccess != nil {
		user.EffectiveAccess.BusinessID = businessID
	}
	return user
}

func directAuthorizationUserWithProvisioningState(app authorization.AppID, role string, state authorization.ProvisioningState, permissions ...string) *authpkg.AuthUser {
	user := directAuthorizationUser(app, role, permissions...)
	user.ProvisioningStatus = &authorization.ProvisioningStatus{
		State:         state,
		EmailVerified: true,
	}
	return user
}

func directAuthorizationUserWithAppState(app authorization.AppID, role string, licensed, hasAccess bool, permissions ...string) *authpkg.AuthUser {
	user := directAuthorizationUser(app, role, permissions...)
	user.EffectiveAccess.Apps[0].Licensed = licensed
	user.EffectiveAccess.Apps[0].HasAccess = hasAccess
	return user
}

func wrongDirectAuthorizationApp(app authorization.AppID) authorization.AppID {
	if app == authorization.AppERP {
		return authorization.AppPOS
	}
	return authorization.AppERP
}

func callDirectAuthorizationCase(ctx context.Context, call func(context.Context) (any, error)) (response any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("direct operation panicked: %v", recovered)
		}
	}()
	return call(ctx)
}

func callDeniedDirectAuthorizationCase(ctx context.Context, t *testing.T, db *sql.DB, call func(context.Context) (any, error)) (any, error) {
	t.Helper()
	before := directDatabaseSnapshot(t, db)
	response, err := callDirectAuthorizationCase(ctx, call)
	after := directDatabaseSnapshot(t, db)
	if len(before) != len(after) {
		t.Fatalf("denied operation changed the database table set")
	}
	for table, beforeSignature := range before {
		if afterSignature := after[table]; afterSignature != beforeSignature {
			t.Fatalf("denied operation changed %s database state from %q to %q", table, beforeSignature, afterSignature)
		}
	}
	return response, err
}

func directDatabaseSnapshot(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	if directAuthorizationProbeTables == nil {
		directAuthorizationProbeUpdatedTables = make(map[string]bool)
		columnRows, err := db.QueryContext(context.Background(), `
			SELECT TABLE_NAME
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'updated_at'`)
		if err != nil {
			t.Fatalf("list direct authorization updated_at columns: %v", err)
		}
		for columnRows.Next() {
			var table string
			if err := columnRows.Scan(&table); err != nil {
				_ = columnRows.Close()
				t.Fatalf("scan direct authorization updated_at table: %v", err)
			}
			directAuthorizationProbeUpdatedTables[table] = true
		}
		if err := columnRows.Err(); err != nil {
			_ = columnRows.Close()
			t.Fatalf("read direct authorization updated_at columns: %v", err)
		}
		if err := columnRows.Close(); err != nil {
			t.Fatalf("close direct authorization updated_at columns: %v", err)
		}

		rows, err := db.QueryContext(context.Background(), `
			SELECT TABLE_NAME
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
			ORDER BY TABLE_NAME`)
		if err != nil {
			t.Fatalf("list direct authorization probe tables: %v", err)
		}
		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err != nil {
				_ = rows.Close()
				t.Fatalf("scan direct authorization probe table: %v", err)
			}
			directAuthorizationProbeTables = append(directAuthorizationProbeTables, table)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatalf("read direct authorization probe tables: %v", err)
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close direct authorization probe tables: %v", err)
		}
	}

	parts := make([]string, 0, len(directAuthorizationProbeTables))
	for _, table := range directAuthorizationProbeTables {
		identifier := strings.ReplaceAll(table, "`", "``")
		literal := strings.ReplaceAll(table, "'", "''")
		updatedAt := "''"
		if directAuthorizationProbeUpdatedTables[table] {
			updatedAt = "COALESCE(DATE_FORMAT(MAX(updated_at), '%Y-%m-%d %H:%i:%s.%f'), '')"
		}
		parts = append(parts, fmt.Sprintf("SELECT '%s' AS table_name, CONCAT(COUNT(*), '|', %s) AS state_signature FROM `%s`", literal, updatedAt, identifier))
	}
	rows, err := db.QueryContext(context.Background(), strings.Join(parts, " UNION ALL "))
	if err != nil {
		t.Fatalf("snapshot direct authorization database: %v", err)
	}
	defer rows.Close()
	snapshot := make(map[string]string, len(directAuthorizationProbeTables))
	for rows.Next() {
		var table string
		var signature string
		if err := rows.Scan(&table, &signature); err != nil {
			t.Fatalf("scan direct authorization database snapshot: %v", err)
		}
		snapshot[table] = signature
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read direct authorization database snapshot: %v", err)
	}
	return snapshot
}

func assertDeniedDirectAuthorizationResponse(t *testing.T, response any, err error) {
	t.Helper()
	if response != nil && directResponseContainsData(response) {
		t.Fatalf("denied operation returned authorized data in a response of type %T", response)
	}
	if err == nil || err.Error() == "" {
		t.Fatal("denied operation returned no response error")
	}
}

func directResponseContainsData(value any) bool {
	return directValueContainsData(reflect.ValueOf(value), make(map[uintptr]bool))
}

func directValueContainsData(value reflect.Value, visited map[uintptr]bool) bool {
	if !value.IsValid() {
		return false
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false
		}
		if value.Kind() == reflect.Pointer {
			pointer := value.Pointer()
			if pointer != 0 && visited[pointer] {
				return false
			}
			if pointer != 0 {
				visited[pointer] = true
			}
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return value.Float() != 0
	case reflect.String:
		return value.String() != ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return value.Len() > 0
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			if field.CanInterface() && directValueContainsData(field, visited) {
				return true
			}
		}
	}
	return false
}

func directAuthorizationStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func openDirectAuthorizationProbeDB(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("DB_HOST") == "" {
		t.Fatal("direct authorization evidence requires DB_HOST and the seeded onion_test database")
	}
	return OpenDB(t)
}
