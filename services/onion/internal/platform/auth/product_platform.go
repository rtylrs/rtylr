package auth

import (
	"context"
	"net/http"
	"strings"
)

type ProductPlatform string

const (
	ProductPlatformHeader                  = "X-Product-Platform"
	ProductPlatformUnknown ProductPlatform = "unknown"
	ProductPlatformRtylr   ProductPlatform = "rtylr"
	ProductPlatformVoxire  ProductPlatform = "voxire"
)

type productPlatformCtxKey struct{}

func ResolveProductPlatform(r *http.Request) ProductPlatform {
	if r == nil {
		return ProductPlatformUnknown
	}

	if headerPlatform := normalizeProductPlatform(r.Header.Get(ProductPlatformHeader)); headerPlatform != ProductPlatformUnknown {
		return headerPlatform
	}

	if headerPlatform := normalizeProductPlatform(r.Header.Get("Product-Platform")); headerPlatform != ProductPlatformUnknown {
		return headerPlatform
	}

	candidates := []string{
		r.Header.Get("Origin"),
		r.Host,
		r.URL.Host,
		r.Referer(),
	}

	for _, candidate := range candidates {
		platform := inferProductPlatform(candidate)
		if platform != ProductPlatformUnknown {
			return platform
		}
	}

	return ProductPlatformUnknown
}

func ProductPlatformCtx(ctx context.Context) ProductPlatform {
	platform, ok := ctx.Value(productPlatformCtxKey{}).(ProductPlatform)
	if !ok {
		return ProductPlatformUnknown
	}

	return platform
}

func WithProductPlatform(ctx context.Context, platform ProductPlatform) context.Context {
	return context.WithValue(ctx, productPlatformCtxKey{}, platform)
}

func normalizeProductPlatform(raw string) ProductPlatform {
	switch ProductPlatform(strings.ToLower(strings.TrimSpace(raw))) {
	case ProductPlatformRtylr:
		return ProductPlatformRtylr
	case ProductPlatformVoxire:
		return ProductPlatformVoxire
	default:
		return ProductPlatformUnknown
	}
}

func inferProductPlatform(candidate string) ProductPlatform {
	candidate = strings.ToLower(candidate)

	switch {
	case strings.Contains(candidate, "rtylr"):
		return ProductPlatformRtylr
	case strings.Contains(candidate, "voxire"), strings.Contains(candidate, "voxelect"):
		return ProductPlatformVoxire
	default:
		return ProductPlatformUnknown
	}
}
