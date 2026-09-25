package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type leaveAccrualService interface {
	RunMonthlyAccrual(context.Context) error
}

func Register(_ context.Context, registry *jobs.Registry, service leaveAccrualService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "hr_leave.monthly_accrual",
		Description: "Run monthly HR leave accruals.",
		Schedule:    jobs.ScheduleFromEnv("HR_LEAVE_MONTHLY_ACCRUAL_SCHEDULE", "0 0 1 * *"),
		Handler:     service.RunMonthlyAccrual,
	})
}
