package eventbus

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	"github.com/voxire/onion/internal/platform/logger"
)

// JobName identifies the fan-out job in the registry and in RUN_JOB.
const JobName = "eventbus.dispatch_events"

// defaultSchedule runs the fan-out every minute. Expansion is cheap, bounded by
// the batch size, and does nothing at all when the outbox is empty, so the cost
// of a frequent pass is a single indexed read. The delay it sets is the floor
// on how long a subscriber waits to hear about anything.
const defaultSchedule = "* * * * *"

// RegisterJobs adds the fan-out job to a registry.
//
// The scheduler holds a MySQL lock per job, so several workers can run this
// without expanding the same event twice. That is belt and braces rather than
// the guarantee: the unique keys on both queues are what actually make a double
// expansion harmless.
func RegisterJobs(registry *platformjobs.Registry, dispatcher *Dispatcher) {
	if registry == nil || dispatcher == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        JobName,
		Description: "Expand platform events into webhook delivery and automation run queues.",
		Schedule:    platformjobs.ScheduleFromEnv("EVENT_DISPATCHER_SCHEDULE", defaultSchedule),
		Handler: func(ctx context.Context) error {
			dispatched, err := dispatcher.DispatchBatch(ctx)
			if err != nil {
				return err
			}
			// Only worth a line when there was something to do; an empty
			// outbox is the normal state and should not fill the log.
			if dispatched > 0 {
				logger.Ctx(ctx).Info("eventbus: expanded events",
					logger.Int("events", dispatched))
			}
			return nil
		},
	})
}
