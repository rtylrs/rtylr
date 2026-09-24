package config

import (
	"net"
	"os"
	"strings"
)

const TrustedProxyCIDRsEnvVar = "TRUSTED_PROXY_CIDRS"

// Cloudflare publishes these proxy ranges for the edge networks that overwrite
// CF-Connecting-IP. They are the safe default for the hosted deployment; a
// deployment with a different edge can provide a comma-separated CIDR list.
// An explicit "none" disables header trust for a directly exposed service.
var defaultTrustedProxyCIDRs = []string{
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",
	"2400:cb00::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
}

// GetTrustedProxyCIDRs returns the configured networks whose peer address is
// allowed to supply CF-Connecting-IP. Invalid entries are ignored rather than
// widening trust; an explicit "none" returns an empty list.
func GetTrustedProxyCIDRs() []string {
	raw := strings.TrimSpace(os.Getenv(TrustedProxyCIDRsEnvVar))
	if strings.EqualFold(raw, "none") {
		return nil
	}
	if raw == "" {
		return append([]string(nil), defaultTrustedProxyCIDRs...)
	}

	entries := strings.Split(raw, ",")
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry = strings.TrimSpace(entry); entry != "" {
			out = append(out, entry)
		}
	}
	return out
}

// IsTrustedProxyIP reports whether remoteAddr belongs to a configured trusted
// proxy network. The socket peer is the security boundary; the header alone is
// never sufficient.
func IsTrustedProxyIP(remoteAddr string) bool {
	host := remoteAddrHost(remoteAddr)
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}

	for _, rawCIDR := range GetTrustedProxyCIDRs() {
		_, network, err := net.ParseCIDR(rawCIDR)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteAddrHost(remoteAddr string) string {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return strings.Trim(remoteAddr, "[]")
}
