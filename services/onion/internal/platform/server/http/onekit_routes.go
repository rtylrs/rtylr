package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// MountOneKit attaches the shared oneKit ServeMux at the generated contract
// paths, such as /hub/user/v1/users/me. Chi's specific routes continue to take
// precedence over this wildcard, allowing generated and platform routes to
// coexist without a separate deployment prefix.
func MountOneKit(router chi.Router, mux *http.ServeMux) {
	router.Handle("/*", mux)
}
