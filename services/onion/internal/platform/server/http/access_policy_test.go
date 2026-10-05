package http

import (
	"context"
	"testing"
)

type anonymousModule struct{}

func (anonymousModule) GetAnonymousRoutes() []string {
	return []string{"/public"}
}

func TestRouteAccessPolicy(t *testing.T) {
	policy := NewRouteAccessPolicy()
	policy.RegisterAnonymousRoutes("/public")
	policy.RegisterAgentRoutes("test-agent", "/status")

	if !policy.IsAllowedAnonymously("/public") {
		t.Fatal("expected anonymous route to be allowed")
	}

	if !policy.IsAllowedForAgent("test-agent", "/status") {
		t.Fatal("expected agent route to be allowed")
	}
}

func TestRouteAccessPolicyMatchesParameterizedRoutes(t *testing.T) {
	policy := NewRouteAccessPolicy()
	policy.RegisterAnonymousRoutes("/pos/orders/v1/orders/{order_id}/public-status")
	policy.RegisterAgentRoutes("test-agent", "/agent/{id}/status")

	if !policy.IsAllowedAnonymously("/pos/orders/v1/orders/123/public-status") {
		t.Fatal("expected anonymous route pattern to match concrete path")
	}
	if policy.IsAllowedAnonymously("/pos/orders/v1/orders/123/items/public-status") {
		t.Fatal("expected route pattern not to match extra path segment")
	}
	if policy.IsAllowedAnonymously("/pos/orders/v1/orders//public-status") {
		t.Fatal("expected route pattern not to match empty path segment")
	}
	if !policy.IsAllowedForAgent("test-agent", "/agent/abc/status") {
		t.Fatal("expected agent route pattern to match concrete path")
	}
}

func TestAutoRegisterAnonymousRoutes(t *testing.T) {
	policy := NewRouteAccessPolicy()

	AutoRegisterAnonymousRoutes(context.Background(), policy, anonymousModule{})

	if !policy.IsAllowedAnonymously("/public") {
		t.Fatal("expected anonymous route from route provider")
	}
}
