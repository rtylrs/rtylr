package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMatchFlowV1RetiringMatchesEveryRetiringDomain(t *testing.T) {
	// One real route per domain Phase 4 deletes, taken from the generated
	// servers rather than invented, so a contract path change breaks this.
	for _, path := range []string{
		"/flow/workspace/v1/workspaces",
		"/flow/project/v1/projects",
		"/flow/task/v1/tasks",
		"/flow/cycle/v1/cycles",
		"/flow/label/v1/labels",
		"/flow/milestone/v1/milestones",
		"/flow/view/v1/views",
		"/flow/insights/v1/overview",
		"/flow/dashboard/v1/dashboard",
		"/flow/time/v1/work-days",
	} {
		if matchFlowV1Retiring(path) == "" {
			t.Errorf("matchFlowV1Retiring(%q) = \"\", want a prefix", path)
		}
	}
}

// The reason this is an allowlist and not a match on "/v1/": these three Flow
// domains stay v1 for good, because they carry no workspace field. Reporting
// them would make the signal worthless.
func TestMatchFlowV1RetiringIgnoresDomainsThatStayV1(t *testing.T) {
	for _, path := range []string{
		"/flow/notification/v1/notifications",
		"/flow/setting/v1/settings",
		"/flow/profile/v1/flow/profile/42",
	} {
		if got := matchFlowV1Retiring(path); got != "" {
			t.Errorf("matchFlowV1Retiring(%q) = %q, want no match", path, got)
		}
	}
}

func TestMatchFlowV1RetiringIgnoresV2AndUnrelatedRoutes(t *testing.T) {
	for _, path := range []string{
		"/flow/project/v2/projects",
		"/flow/task/v2/tasks",
		"/flow/time/v2/work-days/team",
		"/pos/orders/v1/orders",
		"/hub/user/v1/users/me",
		"/health",
		"",
	} {
		if got := matchFlowV1Retiring(path); got != "" {
			t.Errorf("matchFlowV1Retiring(%q) = %q, want no match", path, got)
		}
	}
}

// A v2 path that merely contains a retiring domain's name must not match: the
// check is a prefix, not a substring.
func TestMatchFlowV1RetiringIsPrefixNotSubstring(t *testing.T) {
	if got := matchFlowV1Retiring("/proxy/flow/project/v1/projects"); got != "" {
		t.Errorf("matched a non-prefix path: %q", got)
	}
}

func TestFlowV1UsageMiddlewarePassesTheRequestThrough(t *testing.T) {
	for _, path := range []string{"/flow/project/v1/projects", "/flow/project/v2/projects"} {
		called := false
		handler := FlowV1UsageMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		// Observability must never change behaviour: matched or not, the request
		// reaches the handler and keeps its status.
		if !called {
			t.Errorf("%s: next handler was not called", path)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", path, rec.Code)
		}
	}
}
