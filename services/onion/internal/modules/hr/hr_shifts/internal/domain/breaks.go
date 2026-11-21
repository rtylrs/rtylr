package domain

import (
	"strings"
	"time"
)

const (
	BreakTypeLunch = "Lunch"
	BreakTypeShort = "Short"
	BreakTypeOther = "Other"
)

type BreakRecord struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	EmployeeID      int64
	StaffID         int64
	StaffName       string
	BreakType       string
	StartedAt       time.Time
	EndedAt         *time.Time
	DurationMinutes *int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ListBreaksParams struct {
	EmployeeID *int64
	Limit      int32
	Offset     int32
}

type StartBreakInput struct {
	EmployeeID int64
	StaffID    int64
	StaffName  string
	BreakType  string
}

func NormalizeBreakType(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return BreakTypeLunch, nil
	}
	switch strings.ToLower(trimmed) {
	case "lunch":
		return BreakTypeLunch, nil
	case "short":
		return BreakTypeShort, nil
	case "other":
		return BreakTypeOther, nil
	default:
		return "", ErrInvalidBreakType
	}
}
