package handlers

import (
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
)

func leaveTypeToProto(lt domain.LeaveTypeRecord) *hrv1.LeaveTypePayload {
	var accrualRule *string
	var requiresDocAfterDays *int32

	if lt.AccrualRule != "" {
		accrualRule = &lt.AccrualRule
	}
	if lt.RequiresDocumentationAfterDays != 0 {
		requiresDocAfterDays = &lt.RequiresDocumentationAfterDays
	}

	return &hrv1.LeaveTypePayload{
		Id:                             lt.ID,
		OwnerId:                        lt.OwnerID,
		Name:                           lt.Name,
		IsPaid:                         lt.IsPaid,
		RequiresApproval:               lt.RequiresApproval,
		AccrualRule:                    accrualRule,
		MaxCarryoverDays:               lt.MaxCarryoverDays,
		MinNoticeDays:                  lt.MinNoticeDays,
		RequiresDocumentationAfterDays: requiresDocAfterDays,
		IsActive:                       lt.IsActive,
		CreatedAt:                      lt.CreatedAt.Unix(),
		UpdatedAt:                      lt.UpdatedAt.Unix(),
	}
}

func leaveRequestToProto(r domain.LeaveRequestRecord) *hrv1.LeaveRequestPayload {
	var notes, rejectionReason *string
	if r.Notes != "" {
		notes = &r.Notes
	}
	if r.RejectionReason != "" {
		rejectionReason = &r.RejectionReason
	}

	return &hrv1.LeaveRequestPayload{
		Id:                   r.ID,
		OwnerId:              r.OwnerID,
		EmployeeId:           r.EmployeeID,
		LeaveTypeId:          r.LeaveTypeID,
		StartDate:            r.StartDate,
		EndDate:              r.EndDate,
		TotalDays:            r.TotalDays,
		Notes:                notes,
		Status:               r.Status,
		CurrentApprovalLevel: r.CurrentApprovalLevel,
		RejectionReason:      rejectionReason,
		CreatedAt:            r.CreatedAt.Unix(),
		UpdatedAt:            r.UpdatedAt.Unix(),
	}
}

func leaveBalanceToProto(b domain.LeaveBalanceRecord) *hrv1.LeaveBalancePayload {
	remaining := b.EntitledDays - b.UsedDays - b.PendingDays

	return &hrv1.LeaveBalancePayload{
		Id:            b.ID,
		OwnerId:       b.OwnerID,
		EmployeeId:    b.EmployeeID,
		LeaveTypeId:   b.LeaveTypeID,
		Year:          b.Year,
		EntitledDays:  b.EntitledDays,
		UsedDays:      b.UsedDays,
		PendingDays:   b.PendingDays,
		RemainingDays: remaining,
		CreatedAt:     b.CreatedAt.Unix(),
		UpdatedAt:     b.UpdatedAt.Unix(),
	}
}
