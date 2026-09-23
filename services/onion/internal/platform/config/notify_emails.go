package config

import "strings"

// DefaultLeadNotifyEmails receives website lead notifications when
// LEAD_NOTIFY_EMAILS is unset. A misconfigured deploy must never mean a lead
// is stored and nobody is told.
var DefaultLeadNotifyEmails = []string{
	"info@voxire.com",
	"sales@voxire.com",
	"mohammad.saab@voxire.com",
	"mohamad.homsi@voxire.com",
	"amounehabed@gmail.com",
}

// DefaultNewsletterNotifyEmails receives newsletter signup notifications when
// NEWSLETTER_NOTIFY_EMAILS is unset. Subscribers are lower priority than
// leads, so only the shared inbox hears about them by default.
var DefaultNewsletterNotifyEmails = []string{"info@voxire.com"}

// ParseEmailList splits a comma-separated list of addresses, trims and
// lower-cases each one, drops empties and duplicates while keeping first
// appearance order, and returns a copy of fallback when nothing survives.
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
		return append([]string(nil), fallback...)
	}
	return out
}
