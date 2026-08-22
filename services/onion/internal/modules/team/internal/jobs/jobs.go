package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type terminatedAccessReconciler interface {
	RunTerminatedAccessReconciliation(context.Context) error
}

// Register wires the daily sweep that deactivates app permissions left active
// for terminated employees (#834). The synchronous revoke happens on the
// termination write path; this job is the durability guarantee.
func Register(_ context.Context, registry *jobs.Registry, service terminatedAccessReconciler) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "team.terminated_access_reconciliation",
		Description: "Deactivate app permissions still active for terminated employees.",
		Schedule:    jobs.ScheduleFromEnv("TEAM_TERMINATED_ACCESS_RECONCILIATION_SCHEDULE", "0 3 * * *"),
		Handler:     service.RunTerminatedAccessReconciliation,
	})
}
