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

type fakeNotifier struct{}

func (fakeNotifier) NotifyFailing(context.Context) error { return nil }

// The notifier is a second job on the same module, so a registration that
// only adds one of the two would leave failing endpoints unreported.
func TestRegisterNotifierAddsTheNotifier(t *testing.T) {
	registry := platformjobs.NewRegistry()
	RegisterNotifier(context.Background(), registry, fakeNotifier{})

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("registered %d jobs, want 1", len(jobs))
	}
	if jobs[0].Name != NotifyDeadJobName {
		t.Errorf("job name = %q, want %q", jobs[0].Name, NotifyDeadJobName)
	}
	if jobs[0].Schedule == "" || jobs[0].Handler == nil {
		t.Error("the notifier job would never run: missing schedule or handler")
	}
}

// Both jobs land in one registry, with distinct names.
func TestBothWebhookJobsRegisterTogether(t *testing.T) {
	registry := platformjobs.NewRegistry()
	Register(context.Background(), registry, fakeSender{})
	RegisterNotifier(context.Background(), registry, fakeNotifier{})

	if len(registry.Jobs()) != 2 {
		t.Fatalf("registered %d jobs, want 2", len(registry.Jobs()))
	}
	if JobName == NotifyDeadJobName {
		t.Error("the two jobs share a name, so one would replace the other")
	}
}

func TestRegisterNotifierIgnoresMissingDependencies(t *testing.T) {
	registry := platformjobs.NewRegistry()
	RegisterNotifier(context.Background(), nil, fakeNotifier{})
	RegisterNotifier(context.Background(), registry, nil)
	if len(registry.Jobs()) != 0 {
		t.Errorf("registered %d jobs without dependencies, want 0", len(registry.Jobs()))
	}
}
