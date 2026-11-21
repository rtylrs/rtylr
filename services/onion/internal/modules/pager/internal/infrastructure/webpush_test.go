package infrastructure

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

func testSender(t *testing.T) *WebPush {
	t.Helper()
	private, public, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	sender, err := NewWebPush(public, private, "mailto:push@example.com")
	if err != nil {
		t.Fatal(err)
	}
	return sender
}

// browserSubscription is a subscription with real browser keys, so the
// payload is actually encrypted, aimed at endpoint.
func browserSubscription(t *testing.T, endpoint string) domain.PushSubscription {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	return domain.PushSubscription{
		Endpoint: endpoint,
		P256dh:   base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		Auth:     base64.RawURLEncoding.EncodeToString(secret),
	}
}

func TestNewWebPushRefusesKeysThatAreNotAVAPIDPair(t *testing.T) {
	private, public, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	for name, keys := range map[string][2]string{
		"empty public":  {"", private},
		"empty private": {public, ""},
		"swapped":       {private, public},
		"not base64":    {"not a key!", private},
	} {
		if _, err := NewWebPush(keys[0], keys[1], "mailto:push@example.com"); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := NewWebPush(public, private, " "); err == nil {
		t.Error("accepted an empty subject")
	}
	// Padded keys, as some generators print them, are fine.
	if _, err := NewWebPush(public+"=", private, "https://example.com/contact"); err != nil {
		t.Errorf("padded key: %v", err)
	}
}

func TestSendSignsEncryptsAndReadsThePushServiceAnswer(t *testing.T) {
	sender := testSender(t)
	var status atomic.Int64
	var headers atomic.Pointer[http.Header]
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Clone()
		headers.Store(&h)
		w.WriteHeader(int(status.Load()))
	}))
	defer service.Close()
	sub := browserSubscription(t, service.URL+"/push/abc")

	status.Store(http.StatusCreated)
	if err := sender.Send(context.Background(), sub, []byte(`{"kind":"ready"}`), "pager-1"); err != nil {
		t.Fatalf("201: %v", err)
	}
	last := *headers.Load()
	if auth := last.Get("Authorization"); !strings.HasPrefix(auth, "vapid t=") || !strings.Contains(auth, "k="+sender.PublicKey()) {
		t.Errorf("Authorization = %q, want a VAPID header with our key", auth)
	}
	if last.Get("Content-Encoding") != "aes128gcm" || last.Get("Urgency") != "high" ||
		last.Get("Topic") != "pager-1" || last.Get("TTL") != "600" {
		t.Errorf("headers = %v", last)
	}

	for _, code := range []int{http.StatusNotFound, http.StatusGone} {
		status.Store(int64(code))
		if err := sender.Send(context.Background(), sub, []byte(`{}`), "pager-1"); !errors.Is(err, domain.ErrPushGone) {
			t.Errorf("%d: got %v, want ErrPushGone", code, err)
		}
	}
	status.Store(http.StatusInternalServerError)
	if err := sender.Send(context.Background(), sub, []byte(`{}`), "pager-1"); err == nil || errors.Is(err, domain.ErrPushGone) {
		t.Errorf("500: got %v, want a plain failure that keeps the subscription", err)
	}
}

func TestSendDoesNotFollowRedirects(t *testing.T) {
	var reached atomic.Bool
	elsewhere := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached.Store(true) }))
	defer elsewhere.Close()
	service := httptest.NewServer(http.RedirectHandler(elsewhere.URL, http.StatusTemporaryRedirect))
	defer service.Close()

	err := testSender(t).Send(context.Background(), browserSubscription(t, service.URL+"/push/abc"), []byte(`{}`), "pager-1")
	if err == nil {
		t.Fatal("a redirect counted as delivered")
	}
	if reached.Load() {
		t.Fatal("the push followed a redirect to another host")
	}
}
