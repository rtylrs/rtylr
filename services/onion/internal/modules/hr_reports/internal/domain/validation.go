package domain

import "errors"

var (
	ErrInvalidReportType = errors.New("invalid report type")
	ErrMissingDateRange  = errors.New("report date range is required")
	ErrDateRangeInverted = errors.New("report date_from must be before date_to")
)

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
	return nil
}
