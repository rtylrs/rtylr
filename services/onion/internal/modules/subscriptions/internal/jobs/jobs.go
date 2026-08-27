package jobs

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type subscriptionService interface {
	ExpireTrialSubscriptions(context.Context) error
}

func Register(_ context.Context, registry *platformjobs.Registry, service subscriptionService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        "subscriptions.expire_trials",
		Description: "Expire subscriptions whose trial period has ended.",
		Schedule:    platformjobs.ScheduleFromEnv("SUBSCRIPTIONS_EXPIRE_TRIALS_SCHEDULE", "0 2 * * *"),
		Handler:     service.ExpireTrialSubscriptions,
	})
}
