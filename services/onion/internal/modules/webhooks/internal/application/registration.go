package application

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
)

const maxWebhookURLLength = 2048

var supportedEvents = map[string]struct{}{
	eventbus.TypeOrderCreated: {},
}

func invalidRegistration(format string, args ...any) error {
	return fmt.Errorf("%w: %s", hubdeveloper.ErrInvalidWebhook, fmt.Sprintf(format, args...))
}

// validateRegistration normalises and checks a tenant-supplied registration.
// Delivery re-checks the resolved address at connect time; this rejects what
// can never be delivered so the tenant hears about it now rather than finding
// a dead-lettered delivery later.
func validateRegistration(rawURL string, events []string) (string, []string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", nil, invalidRegistration("url is required")
	}
	if len(rawURL) > maxWebhookURLLength {
		return "", nil, invalidRegistration("url is longer than %d characters", maxWebhookURLLength)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, invalidRegistration("url is not a valid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", nil, invalidRegistration("url must start with http:// or https://")
	}
	host := parsed.Hostname()
	if host == "" {
		return "", nil, invalidRegistration("url has no host")
	}
	if parsed.User != nil {
		return "", nil, invalidRegistration("url must not contain credentials")
	}
	if isInternalHost(host) {
		return "", nil, invalidRegistration("url must point to a public host")
	}

	cleaned := make([]string, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	for _, event := range events {
		event = strings.TrimSpace(event)
		if _, ok := supportedEvents[event]; !ok {
			return "", nil, invalidRegistration("unsupported event %q", event)
		}
		if _, dup := seen[event]; dup {
			continue
		}
		seen[event] = struct{}{}
		cleaned = append(cleaned, event)
	}
	return rawURL, cleaned, nil
}

func isInternalHost(host string) bool {
	lower := strings.ToLower(strings.TrimSuffix(host, "."))
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return isInternalIP(ip)
	}
	return false
}
