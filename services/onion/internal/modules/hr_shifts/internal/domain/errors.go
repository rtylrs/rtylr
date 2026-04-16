package domain

import "errors"

var (
	ErrShiftNotFound  = errors.New("shift not found")
	ErrRosterConflict = errors.New("shift assignment conflicts with approved leave or another shift")
)
