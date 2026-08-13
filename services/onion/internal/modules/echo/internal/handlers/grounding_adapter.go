package handlers

import (
	"context"

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
)

// analystGroundingWithContext is the transport adapter's only concern here:
// it supplies request-scoped quota accounting to the application grounding
// service. Query planning and evidence assembly remain in application.
func (h *GenerateTextStreamHandler) analystGroundingWithContext(ctx context.Context, question string, businessID, userID int64, systemContext string) (*analystGroundingResult, bool) {
	return echoapp.GroundAnalytics(ctx, h.AiService, h.AnalystService, question, businessID, userID, systemContext, func(tokens int) {
		h.recordUsage(ctx, businessID, userID, tokens)
	})
}
