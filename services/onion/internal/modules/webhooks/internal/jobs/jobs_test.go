package jobs

import (
	"context"
	"testing"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type fakeSender struct{}

func (fakeSender) ProcessDue(context.Context) error { return nil }

// The worker registers this by name and RUN_JOB uses the same name.
func TestRegisterAddsTheSender(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), registry, fakeSender{})

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("registered %d jobs, want 1", len(jobs))
	}
	if jobs[0].Name != JobName {
		t.Errorf("job name = %q, want %q", jobs[0].Name, JobName)
	}
	if jobs[0].Schedule == "" || jobs[0].Handler == nil {
		t.Error("the sender job would never run: missing schedule or handler")
	}
}

func TestRegisterIgnoresMissingDependencies(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), nil, fakeSender{})
	Register(context.Background(), registry, nil)
	if len(registry.Jobs()) != 0 {
		t.Errorf("registered %d jobs without dependencies, want 0", len(registry.Jobs()))
	}
}
