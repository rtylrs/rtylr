package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type audienceSyncService interface {
	SyncResendAudience(context.Context) error
}

func Register(_ context.Context, registry *jobs.Registry, service audienceSyncService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "websitelead.sync_newsletter_audience",
		Description: "Mirror voxire.com newsletter subscribers into the Resend audience (RESEND_AUDIENCE_ID).",
		Schedule:    jobs.ScheduleFromEnv("NEWSLETTER_AUDIENCE_SYNC_SCHEDULE", "*/5 * * * *"),
		Handler:     service.SyncResendAudience,
	})
}
