package handlers

import (
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func overtimeRequestToProto(r domain.OvertimeRequestRecord) *hrv1.OvertimeRequestPayload {
	var approvedBy *int64
	var multiplierApplied *float64

	if r.ApprovedBy != 0 {
		approvedBy = &r.ApprovedBy
	}
	if r.MultiplierApplied != 0 {
		multiplierApplied = &r.MultiplierApplied
	}

	return &hrv1.OvertimeRequestPayload{
		Id:                   r.ID,
		OwnerId:              r.OwnerID,
		EmployeeId:           r.EmployeeID,
		Date:                 r.Date,
		PlannedExtraMinutes:  r.PlannedExtraMinutes,
		Reason:               r.Reason,
		Status:               r.Status,
		CurrentApprovalLevel: r.CurrentApprovalLevel,
		ApprovedBy:           approvedBy,
		MultiplierApplied:    multiplierApplied,
		CreatedAt:            r.CreatedAt.Unix(),
		UpdatedAt:            r.UpdatedAt.Unix(),
	}
}
