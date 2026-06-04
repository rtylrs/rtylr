package handlers

import (
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

func shiftToProto(s domain.ShiftRecord) *hrshiftv1.ShiftPayload {
	return &hrshiftv1.ShiftPayload{
		Id:         s.ID,
		OwnerId:    s.OwnerID,
		LocationId: s.LocationID,
		Name:       s.Name,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
		Color:      s.Color,
		CreatedAt:  s.CreatedAt.Unix(),
		UpdatedAt:  s.UpdatedAt.Unix(),
	}
}

func assignmentToProto(a domain.ShiftAssignmentRecord) *hrshiftv1.ShiftAssignmentPayload {
	return &hrshiftv1.ShiftAssignmentPayload{
		Id:         a.ID,
		OwnerId:    a.OwnerID,
		EmployeeId: a.EmployeeID,
		ShiftId:    a.ShiftID,
		Date:       a.Date,
		Status:     a.Status,
		CreatedAt:  a.CreatedAt.Unix(),
		UpdatedAt:  a.UpdatedAt.Unix(),
	}
}
