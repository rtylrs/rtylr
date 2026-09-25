package application

import (
	"context"

	"github.com/google/uuid"
)

// GenerateReportResult is what the P3 stub returns for a "generate"
// request. The real implementation will flip status to `processing` and
// then `ready` once the background worker finishes; at P3 we return
// queued immediately so the frontend can prove the round-trip.
type GenerateReportResult struct {
	ReportID string  `json:"report_id"`
	URL      *string `json:"url"`
	Status   string  `json:"status"`
}

// QueueRevenueReport accepts a generate-revenue request and returns a
// queued receipt. No DB write, no background job — this is a
// contract-level placeholder until the reports worker ships.
func (s *Service) QueueRevenueReport(_ context.Context, _ int64) GenerateReportResult {
	return GenerateReportResult{
		ReportID: uuid.NewString(),
		URL:      nil,
		Status:   "queued",
	}
}
