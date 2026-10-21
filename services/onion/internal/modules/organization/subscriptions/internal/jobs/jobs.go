package jobs

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type subscriptionService interface {
	ExpireTrialSubscriptions(context.Context) error
	ReconcileAllSubscriptions(context.Context) error
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
	// The seat ledger only advances a billing month when something reconciles
	// it, so this is what turns the last seat period into next month's invoice.
	_ = registry.Register(platformjobs.Job{
		Name:        "subscriptions.reconcile",
		Description: "Roll seat billing periods into the current billing month and re-bill changed seat counts.",
		Schedule:    platformjobs.ScheduleFromEnv("SUBSCRIPTIONS_RECONCILE_SCHEDULE", "0 */15 * * * *"),
		Handler:     service.ReconcileAllSubscriptions,
	})
}
