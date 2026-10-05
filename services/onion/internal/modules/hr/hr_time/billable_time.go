package hr_time

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
)

type InvoiceExistenceProvider = application.InvoiceExistenceProvider

type BillableTimeEntry struct {
	ID              int64
	EmployeeID      int64
	ProjectID       *int64
	TaskID          *int64
	Date            string
	DurationMinutes int64
	Description     string
	IsBillable      bool
	Status          string
	Source          string
}

type BillableTimePage struct {
	Entries      []BillableTimeEntry
	Total        int64
	TotalMinutes int64
}

func (m *Module) ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int) (BillableTimePage, error) {
	entries, total, totalMinutes, err := m.service.ListUnbilledBillableTimeEntries(ctx, businessID, limit, offset)
	if err != nil {
		return BillableTimePage{}, err
	}
	page := BillableTimePage{Entries: make([]BillableTimeEntry, 0, len(entries)), Total: total, TotalMinutes: totalMinutes}
	for _, entry := range entries {
		page.Entries = append(page.Entries, billableTimeEntry(entry))
	}
	return page, nil
}

func (m *Module) SetInvoiceExistenceProvider(provider InvoiceExistenceProvider) {
	m.service.SetInvoiceExistenceProvider(provider)
}

func billableTimeEntry(entry domain.TimeEntryRecord) BillableTimeEntry {
	result := BillableTimeEntry{
		ID:              entry.ID,
		EmployeeID:      entry.EmployeeID,
		Date:            entry.Date,
		DurationMinutes: int64(entry.DurationMinutes),
		Description:     entry.Description,
		IsBillable:      entry.IsBillable,
		Status:          entry.Status,
		Source:          entry.Source,
	}
	if entry.ProjectID > 0 {
		value := entry.ProjectID
		result.ProjectID = &value
	}
	if entry.TaskID > 0 {
		value := entry.TaskID
		result.TaskID = &value
	}
	return result
}
