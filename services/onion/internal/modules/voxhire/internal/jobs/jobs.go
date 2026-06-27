package jobs

import (
	"context"

	"github.com/voxire/onion/internal/platform/jobs"
)

type cvProcessingService interface {
	ProcessPendingCVs(context.Context) error
}

func Register(_ context.Context, registry *jobs.Registry, service cvProcessingService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(jobs.Job{
		Name:        "voxhire.process_cvs",
		Description: "Extract searchable text from pending VoxHire CV uploads.",
		Schedule:    jobs.ScheduleFromEnv("VOXHIRE_PROCESS_CVS_SCHEDULE", "*/10 * * * *"),
		Handler:     service.ProcessPendingCVs,
	})
}
