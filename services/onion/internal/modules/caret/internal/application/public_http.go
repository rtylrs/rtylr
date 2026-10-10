package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

// maxPublicResponseBytes bounds any public API or page body Caret reads.
const maxPublicResponseBytes = 2 << 20

// errPrivateAddress is returned when a public fetch resolves to an address
// inside a private network. Caret reads the public internet only.
var errPrivateAddress = errors.New("that address is not on the public internet")

// NewPublicHTTPClient dials public internet addresses only, so a URL from a
// model or a web page can't reach the VPC, the metadata service or localhost.
// Redirects are followed a few times and re-checked on every dial.
func NewPublicHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, Control: publicDialControl}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: timeout,
		MaxIdleConns:          16,
		IdleConnTimeout:       30 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "https" && req.URL.Scheme != "http" {
				return errors.New("redirect to an unsupported scheme")
			}
			return nil
		},
	}
}

func publicDialControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return err
	}
	if !isPublicAddr(ip) {
		return errPrivateAddress
	}
	return nil
}

func isPublicAddr(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"), // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("2001:db8::/32"),
}

// getPublicJSON reads a JSON document from a fixed public API.
func getPublicJSON(ctx context.Context, client *http.Client, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.New("could not build the request")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", caretUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("the service did not respond")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("the service returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPublicResponseBytes+1))
	if err != nil || len(body) > maxPublicResponseBytes {
		return errors.New("the response was too large")
	}
	if err := json.Unmarshal(body, out); err != nil {
		return errors.New("the service returned an invalid response")
	}
	return nil
}

var errNotFound = errors.New("nothing was found")

const caretUserAgent = "rtylr-caret/1.0 (+https://rtylr.com)"
