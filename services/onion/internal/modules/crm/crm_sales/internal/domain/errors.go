package domain

import (
	"errors"
	"fmt"
)

var (
	ErrLeadNotFound     = errors.New("lead not found")
	ErrDealNotFound     = errors.New("deal not found")
	ErrPipelineNotFound = errors.New("pipeline not found")
	ErrStageNotFound    = errors.New("stage not found")
	ErrTaskNotFound     = errors.New("task not found")
	ErrActivityNotFound = errors.New("activity not found")
	ErrContractNotFound = errors.New("contract not found")
	ErrImportNotFound   = errors.New("import not found")

	ErrInvalidAssignee = errors.New("assignee is not a member of this business")
	ErrInvalidCursor   = errors.New("invalid cursor")
	// ErrLeadConverted guards the hand-off: a qualified lead's work continues
	// on its deal, so it cannot be deleted, re-converted or have its status
	// changed directly.
	ErrLeadConverted   = errors.New("lead is already converted")
	ErrDealHasContract = errors.New("deal has a contract")
	ErrPipelineInUse   = errors.New("pipeline still has deals")
	ErrStageInUse      = errors.New("stage still has deals")
	ErrImportClosed    = errors.New("import is no longer running")
)

// ValidationError is a rejected input the caller can fix; its message is
// safe to show.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func invalid(format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// IsValidation reports whether err is a ValidationError.
func IsValidation(err error) bool {
	var target *ValidationError
	return errors.As(err, &target)
}
