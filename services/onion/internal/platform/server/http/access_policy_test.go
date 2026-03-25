package http

import (
	"context"
	"testing"
)

type anonymousModule struct{}

func (anonymousModule) GetAnonymousRoutes() []string {
	return []string{"/public"}
}

type platformContainer struct{}

func (platformContainer) ConfigureAccessPolicy(_ context.Context, policy *RouteAccessPolicy) {
	policy.RegisterAgentRoutes("test-agent", "/status")
}

type rootModules struct {
	Shared   *anonymousModule
	Platform *platformContainer
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

func TestModuleRegistryConfiguresAccessPolicy(t *testing.T) {
	modules := &rootModules{
		Shared:   &anonymousModule{},
		Platform: &platformContainer{},
	}
	policy := NewRouteAccessPolicy()

	AutoRegisterAnonymousRoutes(context.Background(), modules, policy)
	AutoConfigureAccessPolicy(context.Background(), modules, policy)

	if !policy.IsAllowedAnonymously("/public") {
		t.Fatal("expected anonymous route from module registry")
	}

	if !policy.IsAllowedForAgent("test-agent", "/status") {
		t.Fatal("expected agent route from module registry")
	}
}
