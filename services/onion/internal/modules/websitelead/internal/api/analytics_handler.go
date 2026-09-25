package api

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

func AnalyticsEventRoutes() []string { return []string{"/analytics/event"} }

type AnalyticsEventHandler struct{}

// AnalyticsEventRequest mirrors the AnalyticsEventPayload sent by the
// Voxire website frontend. All fields are optional — we log whatever
// arrives and return 200 so the frontend try/catch never fires.
type AnalyticsEventRequest struct {
	Event    string                 `json:"event"`
	Page     string                 `json:"page"`
	JobID    interface{}            `json:"job_id"`
	JobTitle string                 `json:"job_title"`
	Extra    map[string]interface{} `json:"-"`
}

type AnalyticsEventResponse struct {
	Logged bool `json:"logged"`
}

func (h *AnalyticsEventHandler) Serve(ctx context.Context, req *AnalyticsEventRequest) (AnalyticsEventResponse, error) {
	logger.Ctx(ctx).Info("website analytics event",
		logger.Any("event", req.Event),
		logger.Any("page", req.Page),
		logger.Any("job_id", req.JobID),
		logger.Any("job_title", req.JobTitle),
	)
	return AnalyticsEventResponse{Logged: true}, nil
}
