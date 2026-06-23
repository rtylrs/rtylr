package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type payrollPeriodService interface {
	NotifyPayrollPeriodOpen(context.Context) error
}

func Register(_ context.Context, registry *jobs.Registry, service payrollPeriodService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "hr_payroll.period_open_notifications",
		Description: "Notify eligible companies that a payroll period is open.",
		Schedule:    jobs.ScheduleFromEnv("HR_PAYROLL_PERIOD_OPEN_NOTIFICATIONS_SCHEDULE", "0 9 * * *"),
		Handler:     service.NotifyPayrollPeriodOpen,
	})
}
