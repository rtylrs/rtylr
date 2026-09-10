package jobs

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

// Job names, used in the registry and in RUN_JOB.
const (
	JobName           = "webhooks.deliver_due"
	NotifyDeadJobName = "webhooks.notify_dead_deliveries"
)

type webhookSender interface {
	ProcessDue(context.Context) error
}

type deadDeliveryNotifier interface {
	NotifyFailing(context.Context) error
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

// RegisterNotifier adds the dead-delivery notifier to a registry.
//
// Hourly rather than every minute. The message is deduplicated to one per
// registration per day, so a faster schedule would only re-run the query; an
// hour is soon enough for an owner to hear their endpoint has stopped and slow
// enough to cost nothing.
func RegisterNotifier(_ context.Context, registry *platformjobs.Registry, notifier deadDeliveryNotifier) {
	if registry == nil || notifier == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        NotifyDeadJobName,
		Description: "Tell registration owners when their endpoint has stopped receiving events.",
		Schedule:    platformjobs.ScheduleFromEnv("WEBHOOKS_NOTIFY_DEAD_SCHEDULE", "0 * * * *"),
		Handler:     notifier.NotifyFailing,
	})
}
