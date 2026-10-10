package observability

import (
	"net/http"
	"strings"
	"sync/atomic"
)

var routeMux atomic.Pointer[http.ServeMux]

// SetRouteMux registers the ServeMux that serves contract and legacy routes,
// so route templates resolve for requests the outer router hands to it.
func SetRouteMux(mux *http.ServeMux) {
	routeMux.Store(mux)
}

// muxRouteTemplate returns the registered pattern the shared mux would match,
// without its method prefix, or "".
func muxRouteTemplate(r *http.Request) string {
	mux := routeMux.Load()
	if mux == nil || r == nil || r.URL == nil {
		return ""
	}
	_, pattern := mux.Handler(r)
	if _, path, ok := strings.Cut(pattern, " "); ok {
		pattern = path
	}
	return pattern
}

// RouteParamNames lists the {name} wildcards of a route template.
func RouteParamNames(template string) []string {
	var names []string
	for {
		start := strings.IndexByte(template, '{')
		if start < 0 {
			return names
		}
		end := strings.IndexByte(template[start:], '}')
		if end < 0 {
			return names
		}
		names = append(names, strings.TrimSuffix(template[start+1:start+end], "..."))
		template = template[start+end+1:]
	}
}
