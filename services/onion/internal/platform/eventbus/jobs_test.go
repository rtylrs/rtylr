package eventbus_test

import (
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/platform/eventbus"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

// The worker registers the fan-out job by name, and RUN_JOB uses that name for
// a one-shot run. A rename that misses one of those is silent, so the name is
// pinned here.
func TestRegisterJobsAddsTheDispatcher(t *testing.T) {
	registry := platformjobs.NewRegistry()
	eventbus.RegisterJobs(registry, eventbus.NewDispatcher(&sql.DB{}))

	var found bool
	for _, job := range registry.Jobs() {
		if job.Name == eventbus.JobName {
			found = true
			if job.Schedule == "" {
				t.Error("the dispatcher job has no schedule, so it would never run")
			}
			if job.Handler == nil {
				t.Error("the dispatcher job has no handler")
			}
		}
	}
	if !found {
		t.Fatalf("registry does not contain %q; the worker would start without fan-out", eventbus.JobName)
	}
}

// Registration must be inert without its dependencies rather than panicking at
// worker start.
func TestRegisterJobsIgnoresMissingDependencies(t *testing.T) {
	registry := platformjobs.NewRegistry()

	eventbus.RegisterJobs(nil, eventbus.NewDispatcher(&sql.DB{}))
	eventbus.RegisterJobs(registry, nil)

	if len(registry.Jobs()) != 0 {
		t.Errorf("registered %d jobs without dependencies, want 0", len(registry.Jobs()))
	}
}
