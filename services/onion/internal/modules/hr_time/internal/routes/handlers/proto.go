package handlers

import (
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
)

func timeEntryToProto(e domain.TimeEntryRecord) *hrv1.TimeEntryPayload {
	var projectID, taskID, categoryID, approvedBy *int64
	var startTime, endTime, rejectionReason *string

	if e.ProjectID != 0 {
		projectID = &e.ProjectID
	}
	if e.TaskID != 0 {
		taskID = &e.TaskID
	}
	if e.CategoryID != 0 {
		categoryID = &e.CategoryID
	}
	if e.ApprovedBy != 0 {
		approvedBy = &e.ApprovedBy
	}
	if e.StartTime != "" {
		startTime = &e.StartTime
	}
	if e.EndTime != "" {
		endTime = &e.EndTime
	}
	if e.RejectionReason != "" {
		rejectionReason = &e.RejectionReason
	}

	return &hrv1.TimeEntryPayload{
		Id:              e.ID,
		OwnerId:         e.OwnerID,
		EmployeeId:      e.EmployeeID,
		ProjectId:       projectID,
		TaskId:          taskID,
		CategoryId:      categoryID,
		Date:            e.Date,
		DurationMinutes: e.DurationMinutes,
		StartTime:       startTime,
		EndTime:         endTime,
		Description:     e.Description,
		IsBillable:      e.IsBillable,
		Status:          e.Status,
		RejectionReason: rejectionReason,
		ApprovedBy:      approvedBy,
		Source:          e.Source,
		CreatedAt:       e.CreatedAt.Unix(),
		UpdatedAt:       e.UpdatedAt.Unix(),
	}
}

func timesheetToProto(ts domain.TimesheetRecord) *hrv1.TimesheetPayload {
	var submittedAt *int64
	if !ts.SubmittedAt.IsZero() {
		unix := ts.SubmittedAt.Unix()
		submittedAt = &unix
	}

	return &hrv1.TimesheetPayload{
		Id:                   ts.ID,
		OwnerId:              ts.OwnerID,
		EmployeeId:           ts.EmployeeID,
		PeriodStart:          ts.PeriodStart,
		PeriodEnd:            ts.PeriodEnd,
		Status:               ts.Status,
		CurrentApprovalLevel: ts.CurrentApprovalLevel,
		SubmittedAt:          submittedAt,
		TotalMinutes:         ts.TotalMinutes,
		CreatedAt:            ts.CreatedAt.Unix(),
		UpdatedAt:            ts.UpdatedAt.Unix(),
	}
}
