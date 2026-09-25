package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractIPOnlyTrustsConfiguredProxyHeader(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "203.0.113.0/24")
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.RemoteAddr = "198.51.100.10:1234"
	r.Header.Set("CF-Connecting-IP", "8.8.8.8")
	if got := extractIP(r); got != "198.51.100.10" {
		t.Fatalf("direct connection trusted spoofed header: %q", got)
	}

	r.RemoteAddr = "203.0.113.10:1234"
	if got := extractIP(r); got != "8.8.8.8" {
		t.Fatalf("configured proxy header was ignored: %q", got)
	}
}

func TestExtractLanguage(t *testing.T) {
	cases := map[string]string{
		"":               "en",
		"en":             "en",
		"en-US,en;q=0.9": "en",
		"fr-FR,fr;q=0.9": "fr",
		"de-DE":          "de",
		"AR-LB":          "ar",
		"  en-GB  ,en":   "en",
	}
	for header, want := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		if header != "" {
			r.Header.Set("Accept-Language", header)
		}
		if got := extractLanguage(r); got != want {
			t.Errorf("extractLanguage(%q) = %q, want %q", header, got, want)
		}
	}
}

func TestExtractDeviceInfo_FromUserAgent(t *testing.T) {
	cases := map[string]string{
		"": "unknown",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X)":                              "desktop",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17 like Mac OS X) AppleWebKit/605": "mobile",
		"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605":        "tablet",
		"Mozilla/5.0 (Linux; Android 14; Pixel) AppleWebKit/537":               "mobile",
		"Googlebot/2.1 (+http://www.google.com/bot.html)":                      "bot",
	}
	for ua, want := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		if ua != "" {
			r.Header.Set("User-Agent", ua)
		}
		if got := extractDeviceInfo(r); got != want {
			t.Errorf("extractDeviceInfo(%q) = %q, want %q", ua, got, want)
		}
	}
}

func TestGetAnonymousDevice(t *testing.T) {
	d := GetAnonymousDevice()
	if !d.IsAnonymous {
		t.Error("expected IsAnonymous=true")
	}
	if d.User != nil {
		t.Error("expected nil User")
	}
	if d.Language != "en" {
		t.Errorf("expected default language 'en', got %q", d.Language)
	}
}

func TestDeviceCtx_FallsBackToAnonymous(t *testing.T) {
	d := DeviceCtx(context.Background())
	if !d.IsAnonymous {
		t.Error("missing context value should return anonymous device")
	}
}

func TestDeviceCtx_FallsBackToAuthUser(t *testing.T) {
	user := &AuthUser{UserID: 7, BusinessID: 3, IP: "4.5.6.7"}
	ctx := context.WithValue(context.Background(), AuthUserKey, user)

	d := DeviceCtx(ctx)
	if d.IsAnonymous {
		t.Fatal("auth-only context should return an authenticated device")
	}
	if d.User == nil || d.User.UserID != 7 || d.User.BusinessID != 3 {
		t.Fatalf("expected auth user to be copied into device, got %+v", d.User)
	}
	if d.IP != "4.5.6.7" {
		t.Fatalf("expected IP from auth user, got %q", d.IP)
	}
	if d.Language != "en" {
		t.Fatalf("expected default language en, got %q", d.Language)
	}
}

func TestDeviceCtxRoundTrip(t *testing.T) {
	d := Device{
		Language:               "fr",
		IsAnonymous:            false,
		IP:                     "1.2.3.4",
		Platform:               "web",
		DeviceUniqueIdentifier: "device-123",
	}
	ctx := d.Ctx(context.Background())
	got := DeviceCtx(ctx)
	if got.Language != "fr" || got.IP != "1.2.3.4" || got.DeviceUniqueIdentifier != "device-123" {
		t.Errorf("round-trip failed: %+v", got)
	}
}

func TestIsAuthenticatedDevice(t *testing.T) {
	if IsAuthenticatedDevice(context.Background()) {
		t.Error("anonymous context should not be authenticated")
	}

	d := Device{IsAnonymous: false, User: &AuthUser{UserID: 1}}
	ctx := d.Ctx(context.Background())
	if !IsAuthenticatedDevice(ctx) {
		t.Error("populated device should be authenticated")
	}
}

func TestGetUserFromDevice(t *testing.T) {
	if _, ok := GetUserFromDevice(context.Background()); ok {
		t.Error("anonymous should return ok=false")
	}

	d := Device{IsAnonymous: false, User: &AuthUser{UserID: 7}}
	ctx := d.Ctx(context.Background())
	user, ok := GetUserFromDevice(ctx)
	if !ok || user.UserID != 7 {
		t.Errorf("expected user 7, got %+v ok=%v", user, ok)
	}
}

func TestGetLanguageFromDevice_DefaultsToEn(t *testing.T) {
	if got := GetLanguageFromDevice(context.Background()); got != "en" {
		t.Errorf("anonymous default should be 'en', got %q", got)
	}
}
