package http

import (
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

// flowV1RetiringPrefixes are the Flow v1 route prefixes the workspace-removal
// initiative's Phase 4 retires, together with flow_workspaces and every
// workspace_id column.
//
// Deliberately an allowlist of the *retiring* domains rather than a match on
// "/v1/". Three Flow v1 domains stay v1 by design because they carry no
// workspace field: notification, setting and profile. Matching all of /v1/
// would report those forever and train everyone to ignore the signal.
var flowV1RetiringPrefixes = []string{
	"/flow/workspace/v1/",
	"/flow/project/v1/",
	"/flow/task/v1/",
	"/flow/cycle/v1/",
	"/flow/label/v1/",
	"/flow/milestone/v1/",
	"/flow/view/v1/",
	"/flow/insights/v1/",
	"/flow/dashboard/v1/",
	"/flow/time/v1/",
}

// FlowV1UsageMiddleware records every call to a Flow v1 route that Phase 4 will
// delete.
//
// It exists to answer one question the removal plan gates itself on: "have the
// v1 workspace-scoped endpoints received zero traffic for a full release
// cycle?" The plan says to confirm that from logs or metrics rather than
// assumption, and until now nothing could: RequestLoggerMiddleware only reports
// 4xx/5xx, so a *successful* v1 call left no trace at all. Grepping logs and
// finding nothing would have been a false negative, not evidence.
//
// Logged at Warn on purpose. The expected volume after Phase 3 is zero, so any
// line here is a finding: something outside apps/flow is still calling a route
// that is about to disappear.
//
// Delete this file together with the v1 handlers in Phase 4. It has no purpose
// once the routes are gone.
func FlowV1UsageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if route := matchFlowV1Retiring(r.URL.Path); route != "" {
			logger.Ctx(r.Context()).Warn("deprecated flow v1 route called",
				logger.String("route_prefix", route),
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("remote_ip", getRemoteIP(r)),
				logger.String("user_agent", r.UserAgent()),
			)
		}
		next.ServeHTTP(w, r)
	})
}

// matchFlowV1Retiring returns the matched prefix, or "" when the path is not a
// retiring Flow v1 route.
//
// Returns the prefix rather than a bool so the log line says which domain was
// called without re-deriving it, which is what makes the output groupable.
func matchFlowV1Retiring(path string) string {
	for _, prefix := range flowV1RetiringPrefixes {
		if strings.HasPrefix(path, prefix) {
			return prefix
		}
	}
	return ""
}
