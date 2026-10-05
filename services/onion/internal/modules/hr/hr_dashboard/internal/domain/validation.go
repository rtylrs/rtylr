package domain

import "errors"

var (
	ErrInvalidWidgetType      = errors.New("invalid widget type")
	ErrInvalidWidgetPosition  = errors.New("widget position values must be non-negative")
	ErrInvalidWidgetDimension = errors.New("widget width and height must be positive")
)

var supportedWidgetTypes = map[string]bool{
	WidgetTypeHeadcount:         true,
	WidgetTypeAttendanceToday:   true,
	WidgetTypeOpenLeaveRequests: true,
	WidgetTypeUpcomingShifts:    true,
	WidgetTypePayrollSummary:    true,
	WidgetTypeOvertimeStats:     true,
	WidgetTypeRecentHires:       true,
	WidgetTypeTasksDue:          true,
}

// IsValidWidgetType returns true when the widget type is recognised.
func IsValidWidgetType(wt string) bool {
	return supportedWidgetTypes[wt]
}

// ValidateWidget checks that a dashboard widget has valid fields.
func ValidateWidget(w *DashboardWidget) error {
	if w.WidgetType != "" && !IsValidWidgetType(w.WidgetType) {
		return ErrInvalidWidgetType
	}
	if w.PositionX < 0 || w.PositionY < 0 {
		return ErrInvalidWidgetPosition
	}
	if w.Width < 0 || w.Height < 0 {
		return ErrInvalidWidgetDimension
	}
	return nil
}
