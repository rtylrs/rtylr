package jobs

import (
	"context"

	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type recurringInvoiceService interface {
	ProcessDue(context.Context) error
}

func Register(_ context.Context, registry *platformjobs.Registry, service recurringInvoiceService) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(platformjobs.Job{
		Name:        "recurring_invoices.generate_due",
		Description: "Generate invoices for due recurring invoice schedules.",
		Schedule:    platformjobs.ScheduleFromEnv("RECURRING_INVOICES_GENERATE_DUE_SCHEDULE", "0 * * * *"),
		Handler:     service.ProcessDue,
	})
}
