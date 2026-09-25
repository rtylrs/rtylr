package domain

import "errors"

var (
	ErrUnsupportedReportType  = errors.New("unsupported report type")
	ErrReportGenerationFailed = errors.New("report generation failed")
	ErrReportTooLarge         = errors.New("report exceeds the maximum interactive row limit")
)
