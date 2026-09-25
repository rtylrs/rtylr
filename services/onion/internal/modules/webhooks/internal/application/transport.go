package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"syscall"
	"time"
)

// SignatureHeader carries the HMAC a subscriber verifies.
//
// Shaped t=<unix seconds>,v1=<hex> and signed over "<timestamp>.<body>" rather
// than the body alone, so a captured request cannot be replayed indefinitely:
// a subscriber that checks the age of t rejects an old one. Signing the body
// alone would make every replay valid forever.
const SignatureHeader = "X-Rtylr-Signature"

// Headers identifying what is being delivered, so a subscriber can dedupe on
// the event id rather than on payload equality.
const (
	EventTypeHeader  = "X-Rtylr-Event"
	EventIDHeader    = "X-Rtylr-Event-Id"
	DeliveryIDHeader = "X-Rtylr-Delivery-Id"
)

// errUndeliverable marks a destination no number of retries can reach, such as
// one resolving to a private address. Wrapped rather than a sentinel so the
// reason reaches last_error, where an operator reads it.
type errUndeliverable struct{ cause error }

func (e errUndeliverable) Error() string { return e.cause.Error() }
func (e errUndeliverable) Unwrap() error { return e.cause }

func undeliverable(format string, args ...any) error {
	return errUndeliverable{cause: fmt.Errorf(format, args...)}
}

// sign computes the value of SignatureHeader for a body.
func sign(secret string, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(body)
	return fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}

// validateDestination rejects a URL before anything is dialled. Cheap checks
// only; the address itself is judged at connect time by refuseInternalAddress,
// which is the one that matters.
func validateDestination(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return undeliverable("registration url is not a url: %v", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return undeliverable("registration url scheme %q is not http or https", parsed.Scheme)
	}
	if parsed.Host == "" {
		return undeliverable("registration url has no host")
	}
	return nil
}

// internalRanges are the addresses the standard library predicates below do
// not already cover, and which must not be reachable from a tenant-supplied
// URL.
var internalRanges = []string{
	"0.0.0.0/8",     // "this network"
	"100.64.0.0/10", // carrier-grade NAT
	"192.0.0.0/24",  // IETF protocol assignments
	"198.18.0.0/15", // benchmarking
	"240.0.0.0/4",   // reserved
	"::/128",        // unspecified
	"2001:db8::/32", // documentation
}

var parsedInternalRanges = func() []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(internalRanges))
	for _, cidr := range internalRanges {
		if _, network, err := net.ParseCIDR(cidr); err == nil {
			nets = append(nets, network)
		}
	}
	return nets
}()

// isInternalIP reports whether an address belongs to the infrastructure rather
// than the internet.
func isInternalIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return true
	}
	for _, network := range parsedInternalRanges {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// refuseInternalAddress is a net.Dialer Control hook, so it runs after DNS
// resolution with the address actually about to be connected to.
//
// That placement is the point. Checking the hostname, or resolving separately
// and then dialling, both lose to DNS rebinding: a name that answers with a
// public address when validated and 169.254.169.254 when dialled. Judging the
// resolved IP at connect time closes that, and it applies to every address a
// multi-record name offers.
//
// Without this, any tenant could register a webhook pointing at cloud metadata
// or an internal service and have the platform fetch it for them.
func refuseInternalAddress(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("refusing unparseable address %q", address)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("refusing unresolved address %q", host)
	}
	if isInternalIP(ip) {
		return fmt.Errorf("refusing to connect to internal address %s", ip)
	}
	return nil
}

// newDeliveryClient builds the client used for every outbound delivery.
func newDeliveryClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   refuseInternalAddress,
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: timeout,
			// Connections are not pooled across deliveries: they go to
			// unrelated tenants' hosts and reusing one buys nothing.
			DisableKeepAlives: true,
		},
		// Not followed. A redirect is the other way a public URL reaches a
		// private one, and a subscriber that needs a redirect can register the
		// destination instead.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("redirects are not followed")
		},
	}
}
