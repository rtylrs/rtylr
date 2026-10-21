package infrastructure

import (
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

func leadFromRow(row sqlc.CrmLead) domain.Lead {
	return domain.Lead{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		BusinessID:       row.BusinessID,
		FirstName:        row.FirstName.String,
		LastName:         row.LastName.String,
		JobTitle:         row.JobTitle.String,
		Email:            row.Email.String,
		Phone:            row.Phone.String,
		CompanyName:      row.CompanyName,
		Website:          row.Website.String,
		LinkedIn:         row.Linkedin.String,
		Industry:         row.Industry.String,
		Size:             row.Size.String,
		City:             row.City.String,
		Country:          row.Country.String,
		Source:           row.Source,
		Status:           domain.LeadStatus(row.Status),
		DisqualifyReason: row.DisqualifyReason.String,
		AssignedTo:       int64Ptr(row.AssignedTo),
		Value:            amountPtr(row.Value),
		Billing:          domain.Billing(row.Billing),
		Notes:            row.Notes.String,
		AttemptCount:     int(row.AttemptCount),
		LastActivityAt:   timePtr(row.LastActivityAt),
		NextFollowUp:     dateString(row.NextFollowUp),
		CustomerID:       int64Ptr(row.CustomerID),
		CompanyID:        int64Ptr(row.CompanyID),
		DealID:           int64Ptr(row.DealID),
		ConvertedAt:      timePtr(row.ConvertedAt),
		ImportBatchID:    int64Ptr(row.ImportBatchID),
		CreatedAt:        row.CreatedAt.UTC(),
		UpdatedAt:        row.UpdatedAt.UTC(),
	}
}

func leadsFromRows(rows []sqlc.CrmLead) []domain.Lead {
	leads := make([]domain.Lead, 0, len(rows))
	for _, row := range rows {
		leads = append(leads, leadFromRow(row))
	}
	return leads
}

func dealFromRow(row sqlc.CrmDeal) domain.Deal {
	return domain.Deal{
		ID:                row.ID,
		BusinessID:        row.BusinessID,
		OwnerID:           row.OwnerID,
		PipelineID:        row.PipelineID,
		StageID:           row.StageID,
		Title:             row.Title,
		LeadID:            int64Ptr(row.LeadID),
		CustomerID:        int64Ptr(row.CustomerID),
		CompanyID:         int64Ptr(row.CompanyID),
		ContactName:       row.ContactName,
		CompanyName:       row.CompanyName,
		Value:             amount(row.Value),
		Billing:           domain.Billing(row.Billing),
		ExpectedCloseDate: dateString(row.ExpectedCloseDate),
		AssignedTo:        int64Ptr(row.AssignedTo),
		Source:            row.Source,
		Status:            domain.DealStatus(row.Status),
		LostReason:        row.LostReason.String,
		ClosedAt:          timePtr(row.ClosedAt),
		Notes:             row.Notes.String,
		LastActivityAt:    timePtr(row.LastActivityAt),
		NextFollowUp:      dateString(row.NextFollowUp),
		StageChangedAt:    row.StageChangedAt.UTC(),
		CreatedAt:         row.CreatedAt.UTC(),
		UpdatedAt:         row.UpdatedAt.UTC(),
	}
}

func dealsFromRows(rows []sqlc.CrmDeal) []domain.Deal {
	deals := make([]domain.Deal, 0, len(rows))
	for _, row := range rows {
		deals = append(deals, dealFromRow(row))
	}
	return deals
}

func activityFromRow(row sqlc.CrmTouchpoint) domain.Activity {
	activity := domain.Activity{
		ID:         row.ID,
		BusinessID: row.BusinessID,
		LeadID:     int64Ptr(row.LeadID),
		DealID:     int64Ptr(row.DealID),
		Type:       domain.ActivityType(row.Type),
		Outcome:    domain.CallOutcome(row.Outcome),
		Notes:      row.Notes.String,
		CreatedBy:  row.CreatedBy,
		OccurredAt: row.OccurredAt.UTC(),
		CreatedAt:  row.CreatedAt.UTC(),
	}
	if row.DurationSeconds.Valid {
		seconds := int(row.DurationSeconds.Int32)
		activity.DurationSeconds = &seconds
	}
	return activity
}

func taskFromRow(row sqlc.CrmTask) domain.Task {
	return domain.Task{
		ID:          row.ID,
		LeadID:      int64Ptr(row.LeadID),
		DealID:      int64Ptr(row.DealID),
		Title:       row.Title,
		Type:        domain.TaskType(row.Type),
		DueDate:     row.DueDate.Format(time.DateOnly),
		AssignedTo:  int64Ptr(row.AssignedTo),
		CompletedAt: timePtr(row.CompletedAt),
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt.UTC(),
	}
}

// taskListRow is the shape both task list queries return.
type taskListRow struct {
	task                   sqlc.CrmTask
	leadFirst, leadLast    string
	leadCompany, leadPhone string
	dealTitle              string
}

func taskFromListRow(row taskListRow) domain.Task {
	task := taskFromRow(row.task)
	task.LeadName = joinName(row.leadFirst, row.leadLast)
	task.CompanyName = row.leadCompany
	task.Phone = row.leadPhone
	task.DealTitle = row.dealTitle
	return task
}

func joinName(first, last string) string {
	switch {
	case first == "":
		return last
	case last == "":
		return first
	}
	return first + " " + last
}

func stageFromRow(row sqlc.CrmPipelineStage) domain.Stage {
	return domain.Stage{
		ID:          row.ID,
		PipelineID:  row.PipelineID,
		Name:        row.Name,
		Position:    int(row.Position),
		Probability: int(row.Probability),
		Kind:        domain.StageKind(row.Kind),
	}
}

func contractFromRow(row sqlc.CrmContract) domain.Contract {
	return domain.Contract{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		BusinessID:   row.BusinessID,
		LeadID:       int64Ptr(row.LeadID),
		DealID:       int64Ptr(row.DealID),
		Plan:         row.Plan,
		MRR:          amount(row.Mrr),
		Seats:        int(row.Seats),
		Value:        amount(row.Value),
		Billing:      domain.Billing(row.Billing),
		StartDate:    datePtr(row.StartDate),
		EndDate:      datePtr(row.EndDate),
		CustomerID:   int64Ptr(row.CustomerID),
		SignedAt:     datePtr(row.SignedAt),
		SignedByName: row.SignedByName.String,
		ClosedBy:     int64Ptr(row.ClosedBy),
		Notes:        row.Notes.String,
		CreatedAt:    row.CreatedAt.UTC(),
		UpdatedAt:    row.UpdatedAt.UTC(),
	}
}

func importBatchFromRow(row sqlc.CrmImportBatch) domain.ImportBatch {
	return domain.ImportBatch{
		ID:          row.ID,
		FileName:    row.FileName,
		TotalRows:   int(row.TotalRows),
		Created:     int(row.CreatedCount),
		Updated:     int(row.UpdatedCount),
		Skipped:     int(row.SkippedCount),
		Failed:      int(row.FailedCount),
		Status:      string(row.Status),
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt.UTC(),
		CompletedAt: timePtr(row.CompletedAt),
	}
}

func leadStatuses(statuses []domain.LeadStatus) []sqlc.CrmLeadsStatus {
	if len(statuses) == 0 {
		statuses = domain.LeadStatuses()
	}
	out := make([]sqlc.CrmLeadsStatus, 0, len(statuses))
	for _, status := range statuses {
		out = append(out, sqlc.CrmLeadsStatus(status))
	}
	return out
}

func dealStatuses(statuses []domain.DealStatus) []sqlc.CrmDealsStatus {
	if len(statuses) == 0 {
		statuses = []domain.DealStatus{domain.DealOpen, domain.DealWon, domain.DealLost}
	}
	out := make([]sqlc.CrmDealsStatus, 0, len(statuses))
	for _, status := range statuses {
		out = append(out, sqlc.CrmDealsStatus(status))
	}
	return out
}
