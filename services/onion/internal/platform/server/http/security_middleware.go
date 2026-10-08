package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/slack"
)

type blockedRequest struct {
	IP          string
	UserAgent   string
	Path        string
	Method      string
	Headers     map[string][]string
	QueryParams map[string][]string
	Reason      string
	Timestamp   time.Time
}

type securityBatcher struct {
	requests map[string][]blockedRequest
	mutex    sync.RWMutex
	ticker   *time.Ticker
	done     chan bool
}

var (
	batcher *securityBatcher
	once    sync.Once
)

func getBatcher() *securityBatcher {
	once.Do(func() {
		batcher = &securityBatcher{
			requests: make(map[string][]blockedRequest),
			ticker:   time.NewTicker(30 * time.Second),
			done:     make(chan bool),
		}
		go batcher.flushRoutine()
	})

	return batcher
}

func (b *securityBatcher) addRequest(req blockedRequest) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.requests[req.IP] = append(b.requests[req.IP], req)
}

func (b *securityBatcher) flushRoutine() {
	for {
		select {
		case <-b.ticker.C:
			b.flush()
		case <-b.done:
			b.ticker.Stop()

			return
		}
	}
}

func (b *securityBatcher) flush() {
	b.mutex.Lock()
	snapshot := b.requests
	b.requests = make(map[string][]blockedRequest)
	b.mutex.Unlock()

	if len(snapshot) == 0 {
		return
	}

	// One aggregate record per IP per window. Logging each blocked request
	// individually produced thousands of near-identical lines per flush under
	// scanner traffic, so the per-request record is deliberately not emitted.
	ctx := context.Background()
	for ip, reqs := range snapshot {
		logger.Ctx(ctx).Warn("blocked suspicious requests batch",
			logger.String("ip", ip),
			logger.Int("count", len(reqs)),
		)
		for _, req := range reqs {
			slack.Post(slack.Suspicious, fmt.Sprintf("⚠️ blocked request · %s %s · %s · %s", req.Method, req.Path, req.Reason, req.IP))
		}
	}
}

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitized := make(map[string][]string)

	for key, values := range headers {
		if isSensitiveField(key) {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}

	return sanitized
}

func sanitizeQueryParams(params map[string][]string) map[string][]string {
	sanitized := make(map[string][]string)

	for key, values := range params {
		if isSensitiveField(key) {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}

	return sanitized
}

var maliciousPatterns = []string{
	"/.git/",
	"/.env",
	"/config.json",
	"/_all_dbs",
	"/ecp/",
	"/debug/",
	"/@vite/",
	"/server",
	"/about",
	"/phpmyadmin",
	"/admin",
	"/wp-admin",
	"/wp-login",
	"/wp-includes",
	"/wp-content",
	"/wp-config",
	"/xmlrpc.php",
	"/wlwmanifest.xml",
	"/.htaccess",
	"/.htpasswd",
	"/.well-known/",
	"/api/",
	"/graphql",
	"/swagger",
	"/openapi",
	"/actuator",
	"/info",
	"/pong",
	"/ready",
	"/live",
	"/startup",
	"/shutdown",
	"/restart",
	"/reload",
	"/settings",
	"/setup",
	"/install",
	"/upgrade",
	"/migrate",
	"/seed",
	"/backup",
	"/restore",
	"/export",
	"/import",
	"/download",
	"/static",
	"/assets",
	"/css",
	"/js",
	"/fonts",
	"/media",
	"/public",
	"/private",
	"/secret",
	"/hidden",
	"/internal",
	"/system",
	"/root",
	"/home",
	"/tmp",
	"/var",
	"/etc",
	"/usr",
	"/bin",
	"/sbin",
	"/lib",
	"/lib64",
	"/proc",
	"/sys",
	"/dev",
	"/mnt",
	"/opt",
	"/srv",
	"/data",
	"/log/",
	"/logs/",
	"/temp",
	"/run",
	"/lock/",
	"/pid",
	"/socket",
	"/pipe",
	"/fifo",
	"/null",
	"/zero",
	"/random",
	"/urandom",
	"/full",
	"/tty",
	"/console",
	"/pts",
	"/ptmx",
	"/fd",
	"/stdin",
	"/stdout",
	"/stderr",
}
