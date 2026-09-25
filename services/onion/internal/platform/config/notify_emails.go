package config

import (
	"os"
	"strings"
)

// DefaultLeadNotifyEmails receives hosted website lead notifications when
// LEAD_NOTIFY_EMAILS is unset. Self-hosted deployments must configure their
// own recipients.
var DefaultLeadNotifyEmails = []string{
	"info@voxire.com",
	"sales@voxire.com",
	"mohammad.saab@voxire.com",
	"mohamad.homsi@voxire.com",
	"amounehabed@gmail.com",
}

// DefaultNewsletterNotifyEmails receives hosted newsletter signup notifications
// when NEWSLETTER_NOTIFY_EMAILS is unset. Self-hosted deployments must
// configure their own recipients.
var DefaultNewsletterNotifyEmails = []string{"info@voxire.com"}

// ParseEmailList splits a comma-separated list of addresses, trims and
// lower-cases each one, drops empties and duplicates while keeping first
// appearance order, and returns a copy of fallback when nothing survives.
// Self-hosted deployments never inherit or explicitly select Voxire's
// notification addresses.
func ParseEmailList(raw string, fallback []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, part := range strings.Split(raw, ",") {
		addr := strings.ToLower(strings.TrimSpace(part))
		if addr == "" {
			continue
		}
		if _, dup := seen[addr]; dup {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	if len(out) == 0 {
		out = append([]string(nil), fallback...)
	}
	if !selfHostedDeployment() {
		return out
	}

	filtered := out[:0]
	for _, addr := range out {
		if !isVendorNotificationAddress(addr) {
			filtered = append(filtered, addr)
		}
	}
	return filtered
}

func selfHostedDeployment() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("DEPLOYMENT_MODE")))
	mode = strings.ReplaceAll(mode, "-", "_")
	if mode == "self_hosted" || mode == "selfhosted" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SELF_HOSTED"))) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func isVendorNotificationAddress(addr string) bool {
	addr = strings.ToLower(strings.TrimSpace(addr))
	return strings.HasSuffix(addr, "@voxire.com") || addr == "amounehabed@gmail.com"
}
