package jobs

import (
	"context"
	"testing"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type mockService struct{}

func (mockService) ExpireTrialSubscriptions(context.Context) error { return nil }

func TestRegister(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), registry, mockService{})
	job, ok := registry.Get("subscriptions.expire_trials")
	if !ok || job.Handler == nil || job.Schedule == "" {
		t.Fatalf("trial expiry job was not registered: %#v", job)
	}
}
