package domain

import "context"

type Repository interface {
	CreateAttendanceRecord(ctx context.Context, businessID int64, record AttendanceRecordRecord) (*AttendanceRecordRecord, error)
	UpdateAttendanceRecord(ctx context.Context, businessID int64, record AttendanceRecordRecord) (*AttendanceRecordRecord, error)
	GetActiveClockIn(ctx context.Context, businessID, employeeID int64) (*AttendanceRecordRecord, error)
	ListAttendance(ctx context.Context, businessID int64, params ListAttendanceParams) ([]AttendanceRecordRecord, int64, error)
	GetDailyAttendance(ctx context.Context, businessID, locationID int64, date string) ([]AttendanceRecordRecord, error)

	GetAttendanceConfig(ctx context.Context, businessID int64) (*AttendanceConfigRecord, error)

	GetEmployeePINHash(ctx context.Context, businessID, employeeID int64) (string, error)
	SetEmployeePINHash(ctx context.Context, businessID, employeeID int64, hash string) error
	GetLocationCoordinates(ctx context.Context, businessID, locationID int64) (*LocationCoordinates, error)
}
