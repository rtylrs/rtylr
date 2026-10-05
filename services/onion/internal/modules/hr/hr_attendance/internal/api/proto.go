package api

import (
	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func attendanceToProto(r domain.AttendanceRecordRecord) *hrattendancev1.AttendanceRecordPayload {
	var clockOutAt *int64
	var durationMinutes *int32
	var clockInLat, clockInLng *float64

	if !r.ClockOutAt.IsZero() {
		unix := r.ClockOutAt.Unix()
		clockOutAt = &unix
	}
	if !r.ClockOutAt.IsZero() {
		duration := r.DurationMinutes
		durationMinutes = &duration
	}
	if r.ClockInLat != 0 {
		clockInLat = &r.ClockInLat
	}
	if r.ClockInLng != 0 {
		clockInLng = &r.ClockInLng
	}

	return &hrattendancev1.AttendanceRecordPayload{
		Id:              r.ID,
		OwnerId:         r.OwnerID,
		EmployeeId:      r.EmployeeID,
		LocationId:      r.LocationID,
		Date:            r.Date,
		ClockInAt:       r.ClockInAt.Unix(),
		ClockOutAt:      clockOutAt,
		DurationMinutes: durationMinutes,
		ClockInMethod:   r.ClockInMethod,
		ClockInLat:      clockInLat,
		ClockInLng:      clockInLng,
		IsLate:          r.IsLate,
		LateMinutes:     r.LateMinutes,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt.Unix(),
		UpdatedAt:       r.UpdatedAt.Unix(),
	}
}
