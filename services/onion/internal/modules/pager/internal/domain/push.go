package domain

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

// The kinds of push a phone following a ticket receives.
const (
	PushReady  = "ready"
	PushRemind = "remind"
)

// MaxPushSubscriptionsPerTicket bounds what anonymous phones can attach to one
// ticket. A browser that subscribes again refreshes its own row.
const MaxPushSubscriptionsPerTicket = 20

var (
	ErrInvalidPushSubscription  = errors.New("pager push subscription is invalid")
	ErrTooManyPushSubscriptions = errors.New("pager ticket has too many push subscriptions")
	// ErrPushGone is a push service saying the subscription no longer exists.
	ErrPushGone = errors.New("pager push subscription is gone")
)

// PushSubscription is a browser's Web Push subscription for one ticket. It
// identifies a browser install, not a person, and is kept only while the
// ticket is open.
type PushSubscription struct {
	ID         int64
	BusinessID int64
	TicketID   int64
	Endpoint   string
	P256dh     string
	Auth       string
}

// pushServiceHosts are the push services browsers subscribe through. The
// server POSTs to a subscription's endpoint, so any other host is refused
// rather than letting an anonymous caller aim the server at an arbitrary URL.
// A leading dot matches any subdomain.
var pushServiceHosts = []string{
	"fcm.googleapis.com",         // Chrome, Samsung Internet and other Chromium browsers
	".push.services.mozilla.com", // Firefox
	".notify.windows.com",        // Edge on Windows
	"web.push.apple.com",         // Safari on macOS, installed web apps on iOS
}

// ValidatePushSubscription checks a subscription as the browser reported it:
// an endpoint on a known push service and the two keys the payload is
// encrypted with.
func ValidatePushSubscription(endpoint, p256dh, auth string) error {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || !knownPushHost(u.Hostname()) {
		return ErrInvalidPushSubscription
	}
	if key, err := decodeKey(p256dh); err != nil || len(key) != 65 || key[0] != 4 {
		return ErrInvalidPushSubscription
	}
	if secret, err := decodeKey(auth); err != nil || len(secret) != 16 {
		return ErrInvalidPushSubscription
	}
	return nil
}

func knownPushHost(host string) bool {
	host = strings.ToLower(host)
	for _, known := range pushServiceHosts {
		if host == known || strings.HasPrefix(known, ".") && strings.HasSuffix(host, known) {
			return true
		}
	}
	return false
}

// decodeKey reads a key in the URL-safe base64 browsers use, with or without
// padding.
func decodeKey(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(value, "="))
}
