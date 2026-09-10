package jobs

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

// JobName identifies the sender in the registry and in RUN_JOB.
const JobName = "webhooks.deliver_due"

type webhookSender interface {
	ProcessDue(context.Context) error
}

// Register adds the webhook sender to a registry.
//
// Every minute by default. The dispatcher fans out on the same cadence, so a
// subscriber hears about an event within roughly two minutes of it happening.
func Register(_ context.Context, registry *platformjobs.Registry, sender webhookSender) {
	if registry == nil || sender == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        JobName,
		Description: "Deliver queued webhook events to subscriber endpoints.",
		Schedule:    platformjobs.ScheduleFromEnv("WEBHOOKS_DELIVER_DUE_SCHEDULE", "* * * * *"),
		Handler:     sender.ProcessDue,
	})
}
