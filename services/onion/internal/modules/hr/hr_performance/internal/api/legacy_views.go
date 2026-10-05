package api

import (
	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	"time"
)

type performanceConfigRecordPayload struct {
	ID                  int64     `json:"id"`
	OwnerID             int64     `json:"owner_id"`
	BusinessID          int64     `json:"business_id"`
	RatingScaleType     string    `json:"rating_scale_type"`
	RatingMin           int32     `json:"rating_min"`
	RatingMax           int32     `json:"rating_max"`
	RatingLabels        string    `json:"rating_labels"`
	PeerFeedbackEnabled bool      `json:"peer_feedback_enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func payloadPerformanceConfigRecord(value domain.PerformanceConfigRecord) performanceConfigRecordPayload {
	return performanceConfigRecordPayload(value)
}

func payloadPerformanceConfigRecordPointer(value *domain.PerformanceConfigRecord) *performanceConfigRecordPayload {
	if value == nil {
		return nil
	}
	result := payloadPerformanceConfigRecord(*value)
	return &result
}

func payloadPerformanceConfigRecordList(values []domain.PerformanceConfigRecord) []performanceConfigRecordPayload {
	if values == nil {
		return nil
	}
	result := make([]performanceConfigRecordPayload, len(values))
	for i, value := range values {
		result[i] = payloadPerformanceConfigRecord(value)
	}
	return result
}
