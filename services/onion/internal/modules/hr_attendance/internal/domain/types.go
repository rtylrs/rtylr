package domain

import "time"

type AttendanceRecordRecord struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	EmployeeID      int64     `json:"employee_id"`
	LocationID      int64     `json:"location_id"`
	Date            string    `json:"date"`
	ClockInAt       time.Time `json:"clock_in_at"`
	ClockOutAt      time.Time `json:"clock_out_at"`
	DurationMinutes int32     `json:"duration_minutes"`
	ClockInMethod   string    `json:"clock_in_method"`
	ClockInLat      float64   `json:"clock_in_lat"`
	ClockInLng      float64   `json:"clock_in_lng"`
	IsLate          bool      `json:"is_late"`
	LateMinutes     int32     `json:"late_minutes"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AttendanceConfigRecord struct {
	ID                 int64   `json:"id"`
	OwnerID            int64   `json:"owner_id"`
	BusinessID         int64   `json:"business_id"`
	AppEnabled         bool    `json:"app_enabled"`
	QREnabled          bool    `json:"qr_enabled"`
	GPSEnabled         bool    `json:"gps_enabled"`
	PinEnabled         bool    `json:"pin_enabled"`
	FacialEnabled      bool    `json:"facial_enabled"`
	GeofenceRadiusM    int32   `json:"geofence_radius_m"`
	DefaultLocationLat float64 `json:"default_location_lat"`
	DefaultLocationLng float64 `json:"default_location_lng"`
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
