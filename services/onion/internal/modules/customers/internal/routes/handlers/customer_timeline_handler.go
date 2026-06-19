package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CustomerTimelineRoutes() []string {
	return []string{"/crm/customers/timeline"}
}

type GetTimelineHandler struct {
	CRMService *application.Service
}

type GetTimelineRequest struct {
	CustomerID int64 `json:"customer_id"`
}

type TimelineEntryPayload struct {
	ID          int64           `json:"id"`
	CustomerID  int64           `json:"customer_id"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   string          `json:"created_at"`
}

type GetTimelineResponse struct {
	Entries []TimelineEntryPayload `json:"entries"`
}

func (h *GetTimelineHandler) Validate(_ context.Context, req *GetTimelineRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	return nil
}

func (h *GetTimelineHandler) Serve(ctx context.Context, req *GetTimelineRequest) (GetTimelineResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTimelineResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimelineResponse{}, errorModels.PermissionDeniedServerError()
	}

	entries, err := h.CRMService.GetCustomerTimeline(ctx, businessID, req.CustomerID)
	if err != nil {
		return GetTimelineResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get customer timeline",
		}
	}

	payloads := make([]TimelineEntryPayload, len(entries))
	for i, e := range entries {
		payloads[i] = TimelineEntryPayload{
			ID:          e.ID,
			CustomerID:  e.CustomerID,
			Type:        e.Type,
			Description: e.Description,
			Metadata:    json.RawMessage(e.Metadata),
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return GetTimelineResponse{Entries: payloads}, nil
}
