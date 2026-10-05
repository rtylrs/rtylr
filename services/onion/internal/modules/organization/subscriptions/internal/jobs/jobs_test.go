package jobs

import (
	"context"
	"testing"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type mockService struct {
	reconciled int
}

func (mockService) ExpireTrialSubscriptions(context.Context) error { return nil }

func (m *mockService) ReconcileAllSubscriptions(context.Context) error {
	m.reconciled++
	return nil
}

func TestRegister(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), registry, &mockService{})
	job, ok := registry.Get("subscriptions.expire_trials")
	if !ok || job.Handler == nil || job.Schedule == "" {
		t.Fatalf("trial expiry job was not registered: %#v", job)
	}
}

// The seat ledger only moves to a new billing month when this job runs, so a
// missing registration means every month after the first bills zero.
func TestRegister_ReconcilesSeatBilling(t *testing.T) {
	registry := platformjobs.NewRegistry()
	service := &mockService{}
	Register(context.Background(), registry, service)

	job, ok := registry.Get("subscriptions.reconcile")
	if !ok || job.Handler == nil || job.Schedule == "" {
		t.Fatalf("reconcile job was not registered: %#v", job)
	}
	if _, err := registry.Run(context.Background(), "subscriptions.reconcile"); err != nil {
		t.Fatalf("run reconcile job: %v", err)
	}
	if service.reconciled != 1 {
		t.Fatalf("reconcile handler ran %d times, want 1", service.reconciled)
	}
}
