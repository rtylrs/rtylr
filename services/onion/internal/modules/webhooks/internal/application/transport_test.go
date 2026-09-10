package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A registration URL is tenant-supplied. Without this guard any tenant could
// point one at cloud metadata or an internal service and have the platform
// fetch it for them, which is the classic webhook SSRF.
func TestInternalAddressesAreRefused(t *testing.T) {
	refused := []string{
		"127.0.0.1",       // loopback
		"::1",             // loopback v6
		"10.1.2.3",        // RFC1918
		"172.16.0.1",      // RFC1918
		"192.168.1.1",     // RFC1918
		"169.254.169.254", // cloud metadata
		"fd00::1",         // unique local v6
		"fe80::1",         // link local v6
		"0.0.0.0",         // unspecified
		"100.64.0.1",      // carrier-grade NAT
		"192.0.0.1",       // IETF protocol assignments
		"198.18.0.1",      // benchmarking
		"240.0.0.1",       // reserved
		"224.0.0.1",       // multicast
		"::ffff:10.0.0.1", // v4-mapped RFC1918
	}
	for _, address := range refused {
		t.Run(address, func(t *testing.T) {
			if !isInternalIP(net.ParseIP(address)) {
				t.Errorf("%s is treated as deliverable; a tenant could reach it through a webhook", address)
			}
		})
	}

	allowed := []string{"93.184.216.34", "8.8.8.8", "2606:2800:220:1:248:1893:25c8:1946"}
	for _, address := range allowed {
		t.Run("allowed/"+address, func(t *testing.T) {
			if isInternalIP(net.ParseIP(address)) {
				t.Errorf("%s is treated as internal; genuine subscribers would never receive anything", address)
			}
		})
	}
}

// The guard has to sit on the dial, not on the hostname, or a name that
// resolves to a public address when checked and a private one when connected
// walks straight through it. This proves the real client refuses a loopback
// destination even though the URL itself looks ordinary.
func TestTheDeliveryClientRefusesALoopbackDestination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newDeliveryClient(2 * time.Second)
	request, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, server.URL, strings.NewReader("{}"),
	)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	response, err := client.Do(request)
	if err == nil {
		_ = response.Body.Close()
		t.Fatal("the delivery client connected to a loopback address")
	}
	if !isInternalAddressRefusal(err) {
		t.Errorf("connection failed for the wrong reason: %v", err)
	}
}

func TestRefuseInternalAddressAcceptsAPublicAddress(t *testing.T) {
	if err := refuseInternalAddress("tcp", "93.184.216.34:443", nil); err != nil {
		t.Errorf("a public address was refused: %v", err)
	}
	if err := refuseInternalAddress("tcp", "127.0.0.1:443", nil); err == nil {
		t.Error("a loopback address was accepted")
	}
	if err := refuseInternalAddress("tcp", "nonsense", nil); err == nil {
		t.Error("an unparseable address was accepted")
	}
}

// Only http and https, and only a URL with a host.
func TestValidateDestination(t *testing.T) {
	for _, raw := range []string{"file:///etc/passwd", "gopher://x", "ftp://example.test", "https://", "://"} {
		if err := validateDestination(raw); err == nil {
			t.Errorf("%q was accepted as a destination", raw)
		}
	}
	for _, raw := range []string{"https://example.test/hook", "http://example.test:8080/hook"} {
		if err := validateDestination(raw); err != nil {
			t.Errorf("%q was rejected: %v", raw, err)
		}
	}
}

// The signature is the only thing telling a subscriber the payload came from
// us, so it has to be exactly what their library will recompute.
func TestSignatureIsVerifiableAndBoundToTheTimestamp(t *testing.T) {
	const secret = "whsec_example"
	body := []byte(`{"id":"1","type":"order_created"}`)
	timestamp := int64(1757500000)

	header := sign(secret, timestamp, body)

	wantPrefix := fmt.Sprintf("t=%d,v1=", timestamp)
	if !strings.HasPrefix(header, wantPrefix) {
		t.Fatalf("signature header = %q, want it to start %q", header, wantPrefix)
	}

	// What a subscriber does.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", timestamp)))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))

	got := strings.TrimPrefix(header, wantPrefix)
	if !hmac.Equal([]byte(got), []byte(want)) {
		t.Errorf("signature = %s, subscriber would compute %s", got, want)
	}

	// Signed over timestamp and body together, so a replay with a different
	// timestamp does not verify.
	if sign(secret, timestamp+1, body) == header {
		t.Error("the signature does not cover the timestamp, so a captured request replays forever")
	}
	if sign(secret+"x", timestamp, body) == header {
		t.Error("the signature does not depend on the secret")
	}
}
