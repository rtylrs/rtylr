package jobs

import (
	"context"
	"testing"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type fakeExecutor struct{}

func (fakeExecutor) ProcessDue(context.Context) error { return nil }

// The worker registers this by name and RUN_JOB uses the same name, so a
// rename that misses one is silent.
func TestRegisterAddsTheExecutor(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), registry, fakeExecutor{})

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("registered %d jobs, want 1", len(jobs))
	}
	if jobs[0].Name != JobName {
		t.Errorf("job name = %q, want %q", jobs[0].Name, JobName)
	}
	if jobs[0].Schedule == "" || jobs[0].Handler == nil {
		t.Error("the executor job would never run: missing schedule or handler")
	}
}

// Inert without dependencies rather than panicking at worker start.
func TestRegisterIgnoresMissingDependencies(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), nil, fakeExecutor{})
	Register(context.Background(), registry, nil)
	if len(registry.Jobs()) != 0 {
		t.Errorf("registered %d jobs without dependencies, want 0", len(registry.Jobs()))
	}
}
