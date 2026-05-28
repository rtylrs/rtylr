package domain

// Clock-in method values.
const (
	ClockInMethodPin    = "pin"
	ClockInMethodQR     = "qr"
	ClockInMethodApp    = "app"
	ClockInMethodFacial = "facial"
	ClockInMethodGPS    = "gps"
)

// Attendance record status values.
const (
	AttendanceStatusClockedIn  = "clocked_in"
	AttendanceStatusClockedOut = "clocked_out"
	AttendanceStatusAbsent     = "absent"
)
