package domain

import (
	"errors"
	"time"
)

var (
	ErrMissingDateRange  = errors.New("report date range is required")
	ErrDateRangeInverted = errors.New("report date_from must be before date_to")
	ErrDateRangeTooLarge = errors.New("report date range exceeds the maximum interactive window")
)

const MaxInteractiveReportWindow = 366 * 24 * time.Hour

// IsValidReportType returns true when the report type is recognised.
func IsValidReportType(rt string) bool {
	return SupportedReportTypes[rt]
}

// ValidateReportParams checks required fields on a report request.
func ValidateReportParams(p ReportParams) error {
	if p.DateFrom.IsZero() || p.DateTo.IsZero() {
		return ErrMissingDateRange
	}
	if !p.DateTo.After(p.DateFrom) {
		return ErrDateRangeInverted
	}
	if p.DateTo.Sub(p.DateFrom) > MaxInteractiveReportWindow {
		return ErrDateRangeTooLarge
	}
	return nil
}
