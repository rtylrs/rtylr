package domain

import "time"

type AttendanceRecordRecord struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	EmployeeID      int64
	LocationID      int64
	Date            string
	ClockInAt       time.Time
	ClockOutAt      time.Time
	DurationMinutes int32
	ClockInMethod   string
	ClockInLat      float64
	ClockInLng      float64
	IsLate          bool
	LateMinutes     int32
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AttendanceConfigRecord struct {
	ID                 int64
	OwnerID            int64
	BusinessID         int64
	AppEnabled         bool
	QREnabled          bool
	GPSEnabled         bool
	PinEnabled         bool
	FacialEnabled      bool
	GeofenceRadiusM    int32
	DefaultLocationLat float64
	DefaultLocationLng float64
}

type ListAttendanceParams struct {
	EmployeeID int64
	DateFrom   string
	DateTo     string
	Limit      int32
	Offset     int32
}

type LocationCoordinates struct {
	Latitude  float64
	Longitude float64
	HasCoords bool
}
