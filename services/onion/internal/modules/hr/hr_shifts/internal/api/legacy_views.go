package api

import (
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
)

type conflictInfoPayload struct {
	EmployeeID   int64  `json:"employee_id"`
	Date         string `json:"date"`
	ConflictType string `json:"conflict_type"`
	Description  string `json:"description"`
}

func payloadConflictInfo(value domain.ConflictInfo) conflictInfoPayload {
	return conflictInfoPayload(value)
}

func payloadConflictInfoPointer(value *domain.ConflictInfo) *conflictInfoPayload {
	if value == nil {
		return nil
	}
	result := payloadConflictInfo(*value)
	return &result
}

func payloadConflictInfoList(values []domain.ConflictInfo) []conflictInfoPayload {
	if values == nil {
		return nil
	}
	result := make([]conflictInfoPayload, len(values))
	for i, value := range values {
		result[i] = payloadConflictInfo(value)
	}
	return result
}
