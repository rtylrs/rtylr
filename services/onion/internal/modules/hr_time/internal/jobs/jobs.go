package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type unsubmittedTimesheetReminderService interface {
	RunUnsubmittedTimesheetReminder(context.Context) error
}

func Register(_ context.Context, registry *jobs.Registry, service unsubmittedTimesheetReminderService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "hr_time.unsubmitted_timesheet_reminder",
		Description: "Scan HR time entries for unsubmitted timesheets.",
		Schedule:    jobs.ScheduleFromEnv("HR_TIME_UNSUBMITTED_TIMESHEET_REMINDER_SCHEDULE", "0 9 * * MON"),
		Handler:     service.RunUnsubmittedTimesheetReminder,
	})
}
