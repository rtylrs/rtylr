package jobs

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

// JobName identifies the executor in the registry and in RUN_JOB.
const JobName = "automations.process_due_runs"

type automationExecutor interface {
	ProcessDue(context.Context) error
}

// Register adds the automation executor to a registry.
//
// Every minute by default. The dispatcher fans out on the same cadence, so an
// automation fires within roughly two minutes of the event that triggered it.
// Bringing that down means running both more often, not just this one.
func Register(_ context.Context, registry *platformjobs.Registry, executor automationExecutor) {
	if registry == nil || executor == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        JobName,
		Description: "Execute queued automation runs produced by the event dispatcher.",
		Schedule:    platformjobs.ScheduleFromEnv("AUTOMATIONS_PROCESS_DUE_RUNS_SCHEDULE", "* * * * *"),
		Handler:     executor.ProcessDue,
	})
}
