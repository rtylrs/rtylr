package domain

import "errors"

var (
	ErrReviewCycleNotFound       = errors.New("review cycle not found")
	ErrReviewNotFound            = errors.New("performance review not found")
	ErrReviewAlreadySubmitted    = errors.New("review has already been submitted")
	ErrPerformanceConfigNotFound = errors.New("performance config not found")
	ErrGoalNotFound              = errors.New("review goal not found")
	ErrInvalidRating             = errors.New("rating is invalid for the configured scale")
	ErrCycleNotDraft             = errors.New("review cycle is not in draft status")
	ErrCycleAlreadyOpen          = errors.New("review cycle is already open")
	ErrReviewNotSubmitted        = errors.New("review has not been submitted")
	ErrReviewNotCompleted        = errors.New("review has not been completed")
	ErrEmployeeNotFound          = errors.New("employee not found")
	ErrGoalSubjectNotAllowed     = errors.New("caller is not allowed to manage this employee")
	ErrPerformanceAccessDenied   = errors.New("performance management access denied")
)
