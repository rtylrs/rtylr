package application

import (
	"context"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestIsPublicAddrRejectsInternalNetworks(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "10.0.3.4", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.64.1.1", "::1", "fe80::1", "fd00::1", "0.0.0.0", "::ffff:10.0.0.1"} {
		if isPublicAddr(netip.MustParseAddr(address)) {
			t.Errorf("%s was treated as public", address)
		}
	}
	for _, address := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111"} {
		if !isPublicAddr(netip.MustParseAddr(address)) {
			t.Errorf("%s was treated as private", address)
		}
	}
}

func TestPublicDialControlBlocksMetadataService(t *testing.T) {
	if err := publicDialControl("tcp4", "169.254.169.254:80", nil); err == nil {
		t.Fatal("the metadata service must not be reachable")
	}
	if err := publicDialControl("tcp4", "1.1.1.1:443", nil); err != nil {
		t.Fatalf("public address rejected: %v", err)
	}
}

func TestGetPublicJSONRetriesOnceOnTransientFailure(t *testing.T) {
	previous := publicRetryDelay
	publicRetryDelay = time.Millisecond
	defer func() { publicRetryDelay = previous }()
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("busy"))}, nil
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, nil
	})}
	var out map[string]bool
	if err := getPublicJSON(context.Background(), client, "https://api.example/x", &out); err != nil || !out["ok"] || calls != 2 {
		t.Fatalf("err=%v out=%v calls=%d", err, out, calls)
	}
	calls = 10 // every later call fails
	client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 429, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	if err := getPublicJSON(context.Background(), client, "https://api.example/x", &out); err == nil || err.Error() != "the service returned status 429" || calls != 12 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestGetPublicJSONCachesRepeatCalls(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"rate":1.5}`))}, nil
	})}
	for i := 0; i < 3; i++ {
		var out map[string]float64
		if err := getPublicJSON(context.Background(), client, "https://api.example/rates?base=USD", &out); err != nil || out["rate"] != 1.5 {
			t.Fatalf("call %d: %v %v", i, out, err)
		}
	}
	if calls != 1 {
		t.Fatalf("repeat calls hit the API %d times", calls)
	}
}
