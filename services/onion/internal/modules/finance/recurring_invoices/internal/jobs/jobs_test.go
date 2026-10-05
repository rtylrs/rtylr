package jobs

import (
	"context"
	"testing"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type fakeService struct{}

func (fakeService) ProcessDue(context.Context) error { return nil }

func TestRegister(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), registry, fakeService{})

	job, ok := registry.Get("recurring_invoices.generate_due")
	if !ok {
		t.Fatal("recurring invoice job was not registered")
	}
	if job.Handler == nil || job.Schedule == "" {
		t.Fatalf("registered job is incomplete: %+v", job)
	}
}
