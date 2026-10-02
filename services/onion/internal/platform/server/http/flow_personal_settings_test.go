package http

import (
	"net/http"
	"testing"
)

// GET and PUT /flow/setting/v1/settings read and write one row keyed on the
// caller's own user id. Mapping them onto flow:settings:manage meant an
// ordinary member could not read or change their own notification preferences,
// and because sidebar_preferences lives in the same row, could not load or save
// their sidebar either. The middleware must not claim a permission for them.
func TestFlowPersonalSettingsAreNotGatedOnABusinessPermission(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		appID, permission, ok := featurePermissionForRequest(method, "/flow/setting/v1/settings")
		if ok {
			t.Errorf("%s /flow/setting/v1/settings maps to %s:%s; personal preferences must not require a business permission",
				method, appID, permission)
		}
	}
}

// The rest of the Flow table is untouched, so a regression that drops more than
// the /settings branch is caught here rather than in production.
func TestFlowFeaturePermissionsStillMapTheBusinessWideRoutes(t *testing.T) {
	cases := []struct {
		method     string
		path       string
		permission string
	}{
		{http.MethodGet, "/flow/task/v2/tasks", "flow:tasks:view"},
		{http.MethodPost, "/flow/task/v2/tasks", "flow:tasks:create"},
		{http.MethodPut, "/flow/task/v2/tasks/1", "flow:tasks:edit"},
		{http.MethodPost, "/flow/task/v2/tasks/1/assign", "flow:tasks:assign"},
		{http.MethodGet, "/flow/project/v2/projects", "flow:projects:view"},
		{http.MethodGet, "/flow/project/v2/projects/1/members", "flow:projects:view"},
		{http.MethodPost, "/flow/project/v2/projects/1/members", "flow:projects:assign"},
		{http.MethodDelete, "/flow/project/v2/projects/1/members/2", "flow:projects:assign"},
		{http.MethodPost, "/flow/project/v2/projects", "flow:projects:create"},
		{http.MethodGet, "/flow/project/v2/managers", "flow:managers:manage"},
		{http.MethodGet, "/flow/team/workload", "flow:team:manage"},
	}

	for _, tc := range cases {
		_, permission, ok := featurePermissionForRequest(tc.method, tc.path)
		if !ok || permission != tc.permission {
			t.Errorf("%s %s = %q,%t; want %q", tc.method, tc.path, permission, ok, tc.permission)
		}
	}
}
