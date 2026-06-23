package http

import (
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

// HostDispatcher routes requests by Host header with a fallback router. It is
// useful for self-host installs where api/auth/dash/pos/erp may live on fully
// custom hostnames instead of predictable subdomains.
type HostDispatcher struct {
	mu       sync.RWMutex
	fallback http.Handler
	hosts    map[string]http.Handler
}

func NewHostDispatcher(fallback chi.Router) *HostDispatcher {
	return &HostDispatcher{
		fallback: fallback,
		hosts:    make(map[string]http.Handler),
	}
}

func (d *HostDispatcher) Register(host string, handler http.Handler) {
	host = normalizeHost(host)
	if host == "" || handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.hosts[host] = handler
}

func (d *HostDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := normalizeHost(r.Host)

	d.mu.RLock()
	handler := d.hosts[host]
	fallback := d.fallback
	d.mu.RUnlock()

	if handler == nil {
		handler = fallback
	}
	if handler == nil {
		http.NotFound(w, r)
		return
	}
	handler.ServeHTTP(w, r)
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}
	if i := strings.LastIndex(host, ":"); i > -1 && !strings.Contains(host[i+1:], "]") {
		host = host[:i]
	}
	return host
}
