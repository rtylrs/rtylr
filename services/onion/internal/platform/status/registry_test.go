package status

import (
	"context"
	"testing"
)

type fakeProvider struct {
	name   string
	status string
}

func (f fakeProvider) GetName() string { return f.name }
func (f fakeProvider) GetStatus(_ context.Context) ComponentStatus {
	return ComponentStatus{Name: f.name, Status: f.status, LastChecked: "now"}
}

func TestRegistry_RegisterAndGetSystemStatus(t *testing.T) {
	r := &Registry{providers: map[string]Provider{}}
	r.RegisterProvider(fakeProvider{name: "db", status: "healthy"})
	r.RegisterProvider(fakeProvider{name: "cache", status: "healthy"})

	got := r.GetSystemStatus(context.Background())
	if got.OverallStatus != "healthy" {
		t.Errorf("expected overall healthy, got %q", got.OverallStatus)
	}
	if len(got.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(got.Components))
	}
}

func TestRegistry_OneUnhealthyTaintsOverall(t *testing.T) {
	r := &Registry{providers: map[string]Provider{}}
	r.RegisterProvider(fakeProvider{name: "db", status: "healthy"})
	r.RegisterProvider(fakeProvider{name: "cache", status: "unhealthy"})

	got := r.GetSystemStatus(context.Background())
	if got.OverallStatus != "unhealthy" {
		t.Errorf("expected overall unhealthy, got %q", got.OverallStatus)
	}
}

func TestRegistry_EmptyIsHealthy(t *testing.T) {
	r := &Registry{providers: map[string]Provider{}}
	got := r.GetSystemStatus(context.Background())
	if got.OverallStatus != "healthy" {
		t.Errorf("expected empty registry to be healthy, got %q", got.OverallStatus)
	}
	if len(got.Components) != 0 {
		t.Errorf("expected no components, got %d", len(got.Components))
	}
}

func TestRegistry_RegisterReplacesByName(t *testing.T) {
	r := &Registry{providers: map[string]Provider{}}
	r.RegisterProvider(fakeProvider{name: "db", status: "healthy"})
	r.RegisterProvider(fakeProvider{name: "db", status: "unhealthy"})

	got := r.GetSystemStatus(context.Background())
	if got.Components["db"].Status != "unhealthy" {
		t.Errorf("expected re-registration to replace, got %q", got.Components["db"].Status)
	}
}

func TestGetRegistry_Singleton(t *testing.T) {
	a := GetRegistry()
	b := GetRegistry()
	if a != b {
		t.Error("GetRegistry should return the same instance")
	}
}
