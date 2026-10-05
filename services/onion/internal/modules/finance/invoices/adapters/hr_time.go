package invoiceadapters

import (
	"context"

	"github.com/voxire/onion/internal/modules/finance/invoices"
	hrtime "github.com/voxire/onion/internal/modules/hr/hr_time"
)

type financeHRBillableTimeAdapter struct {
	module *hrtime.Module
}

func (a financeHRBillableTimeAdapter) ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int) ([]invoices.BillableTimeEntry, int64, int64, error) {
	page, err := a.module.ListUnbilledBillableTimeEntries(ctx, businessID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	entries := make([]invoices.BillableTimeEntry, 0, len(page.Entries))
	for _, entry := range page.Entries {
		entries = append(entries, invoices.BillableTimeEntry{
			ID:              entry.ID,
			EmployeeID:      entry.EmployeeID,
			ProjectID:       entry.ProjectID,
			TaskID:          entry.TaskID,
			Date:            entry.Date,
			DurationMinutes: entry.DurationMinutes,
			Description:     entry.Description,
			IsBillable:      entry.IsBillable,
			Status:          entry.Status,
			Source:          entry.Source,
		})
	}
	return entries, page.Total, page.TotalMinutes, nil
}

func NewBillableTimeReader(module *hrtime.Module) invoices.BillableTimeReader {
	return financeHRBillableTimeAdapter{module: module}
}
