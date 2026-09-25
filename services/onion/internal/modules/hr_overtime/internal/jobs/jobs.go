package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type overtimeDetectionService interface {
	RunDailyDetection(context.Context) error
}

func Register(_ context.Context, registry *jobs.Registry, service overtimeDetectionService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "hr_overtime.daily_detection",
		Description: "Run daily HR overtime detection.",
		Schedule:    jobs.ScheduleFromEnv("HR_OVERTIME_DAILY_DETECTION_SCHEDULE", "0 0 * * *"),
		Handler:     service.RunDailyDetection,
	})
}
