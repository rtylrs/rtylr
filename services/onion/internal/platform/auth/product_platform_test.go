package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubTokenValidator struct{}

func (stubTokenValidator) ValidateToken(context.Context, string) (int64, error) {
	return 1, nil
}

func (stubTokenValidator) ValidateTokenFull(context.Context, string) (*TokenClaims, error) {
	return &TokenClaims{UserID: 1}, nil
}

func TestResolveProductPlatformFromHeader(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "http://localhost/account", http.NoBody)
	req.Header.Set(ProductPlatformHeader, "rtylr")
	req.Host = "api.voxire.com"

	if got := ResolveProductPlatform(req); got != ProductPlatformRtylr {
		t.Fatalf("expected header platform %q, got %q", ProductPlatformRtylr, got)
	}
}

func TestResolveProductPlatformFromOriginFallback(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "http://localhost/account", http.NoBody)
	req.Header.Set("Origin", "https://app.voxire.com")

	if got := ResolveProductPlatform(req); got != ProductPlatformVoxire {
		t.Fatalf("expected origin platform %q, got %q", ProductPlatformVoxire, got)
	}
}

func TestWithProductPlatformRoundTrip(t *testing.T) {
	ctx := WithProductPlatform(context.Background(), ProductPlatformVoxire)

	if got := ProductPlatformCtx(ctx); got != ProductPlatformVoxire {
		t.Fatalf("expected context platform %q, got %q", ProductPlatformVoxire, got)
	}
}

func TestStaticTokenValidatorResolver(t *testing.T) {
	defaultValidator := stubTokenValidator{}
	specificValidator := stubTokenValidator{}
	resolver := NewStaticTokenValidatorResolver(defaultValidator, map[ProductPlatform]TokenValidator{
		ProductPlatformRtylr: specificValidator,
	})

	if got := resolver.ResolveTokenValidator(context.Background(), ProductPlatformRtylr); got != specificValidator {
		t.Fatal("expected platform-specific validator")
	}

	if got := resolver.ResolveTokenValidator(context.Background(), ProductPlatformVoxire); got != defaultValidator {
		t.Fatal("expected default validator fallback")
	}
}
