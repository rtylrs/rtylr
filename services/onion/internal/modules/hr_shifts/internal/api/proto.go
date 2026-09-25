package api

import (
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

// shiftToProto maps a shift row to its contract payload. The attendance fields
// are optional in the contract and are only populated when the row actually
// carries them, so a template row does not emit a wall of zero values.
func shiftToProto(s domain.ShiftRecord) *hrshiftv1.ShiftPayload {
	payload := &hrshiftv1.ShiftPayload{
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

	if s.StaffID != 0 {
		payload.StaffId = &s.StaffID
	}
	if s.DepartmentID != 0 {
		payload.DepartmentId = &s.DepartmentID
	}
	if s.Date != "" {
		payload.Date = &s.Date
	}
	if s.BreakDurationMinutes != 0 {
		payload.BreakDurationMinutes = &s.BreakDurationMinutes
	}
	if s.Status != "" {
		payload.Status = &s.Status
	}
	if !s.ClockInAt.IsZero() {
		clockIn := s.ClockInAt.Unix()
		payload.ClockInAt = &clockIn
	}
	if !s.ClockOutAt.IsZero() {
		clockOut := s.ClockOutAt.Unix()
		payload.ClockOutAt = &clockOut
	}
	if s.HoursWorked != 0 {
		payload.HoursWorked = &s.HoursWorked
	}
	if s.OvertimeHours != 0 {
		payload.OvertimeHours = &s.OvertimeHours
	}
	if s.Notes != "" {
		payload.Notes = &s.Notes
	}

	return payload
}

// availabilityToProto maps an availability row to its contract payload. Exactly
// one of DayOfWeek and SpecificDate is populated, matching the row's kind.
func availabilityToProto(a domain.AvailabilityRecord) *hrshiftv1.AvailabilityPayload {
	payload := &hrshiftv1.AvailabilityPayload{
		Id:          a.ID,
		StaffId:     a.StaffID,
		OwnerId:     a.OwnerID,
		StartTime:   a.StartTime,
		EndTime:     a.EndTime,
		IsAvailable: a.IsAvailable,
		CreatedAt:   a.CreatedAt.Unix(),
		UpdatedAt:   a.UpdatedAt.Unix(),
	}
	if a.DayOfWeek != domain.NoDayOfWeek {
		day := a.DayOfWeek
		payload.DayOfWeek = &day
	}
	if a.SpecificDate != "" {
		date := a.SpecificDate
		payload.SpecificDate = &date
	}
	return payload
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
