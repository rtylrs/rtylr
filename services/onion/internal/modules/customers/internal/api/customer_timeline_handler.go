package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CustomerTimelineRoutes() []string {
	return []string{"/crm/customers/timeline"}
}

type GetTimelineHandler struct {
	CRMService *application.Service
}

type GetTimelineRequest struct {
	CustomerID int64  `json:"customer_id"`
	Cursor     string `json:"cursor"`
	Limit      int    `json:"limit"`
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
	Entries    []TimelineEntryPayload `json:"entries"`
	NextCursor string                 `json:"next_cursor,omitempty"`
}

func (h *GetTimelineHandler) Validate(_ context.Context, req *GetTimelineRequest) error {
	if req.CustomerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	return nil
}

func (h *GetTimelineHandler) Serve(ctx context.Context, req *GetTimelineRequest) (GetTimelineResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTimelineResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTimelineResponse{}, onekiterrors.PermissionDenied()
	}

	cursorCreatedAt, cursorID, err := parseTimelineCursor(req.Cursor)
	if err != nil {
		return GetTimelineResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid timeline cursor")
	}
	page, err := h.CRMService.ListCustomerTimelinePage(ctx, businessID, req.CustomerID, cursorCreatedAt, cursorID, req.Limit)
	if err != nil {
		return GetTimelineResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get customer timeline")
	}

	payloads := make([]TimelineEntryPayload, len(page.Entries))
	for i, e := range page.Entries {
		payloads[i] = TimelineEntryPayload{
			ID:          e.ID,
			CustomerID:  e.CustomerID,
			Type:        e.Type,
			Description: e.Description,
			Metadata:    json.RawMessage(e.Metadata),
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return GetTimelineResponse{Entries: payloads, NextCursor: page.NextCursor}, nil
}

func parseTimelineCursor(cursor string) (*time.Time, int64, error) {
	if cursor == "" {
		return nil, 0, nil
	}
	parts := strings.Split(cursor, ":")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid cursor")
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, 0, err
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return nil, 0, fmt.Errorf("invalid cursor id")
	}
	createdAt := time.Unix(0, nanos).UTC()
	return &createdAt, id, nil
}
