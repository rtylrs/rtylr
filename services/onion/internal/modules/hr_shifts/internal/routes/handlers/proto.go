package handlers

import (
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
)

func shiftToProto(s domain.ShiftRecord) *hrv1.ShiftPayload {
	return &hrv1.ShiftPayload{
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

func assignmentToProto(a domain.ShiftAssignmentRecord) *hrv1.ShiftAssignmentPayload {
	return &hrv1.ShiftAssignmentPayload{
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
