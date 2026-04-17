package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/application"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// -- LIST INBOX --------------------------------------------------------------

func ListInboxRoutes() []string { return []string{"/support/inbox"} }

type ListInboxHandler struct{ Service *application.Service }

type ListInboxRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ListInboxResponse struct {
	Inbox    []domain.InboxItem `json:"inbox"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

func (h *ListInboxHandler) Serve(ctx context.Context, req *ListInboxRequest) (ListInboxResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListInboxResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	items, total, err := h.Service.ListInbox(ctx, device.User.BusinessID, req.Page, req.PageSize)
	if err != nil {
		return ListInboxResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list inbox"}
	}
	return ListInboxResponse{Inbox: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// -- UNREAD COUNT ------------------------------------------------------------

func UnreadCountRoutes() []string { return []string{"/support/inbox/unread-count"} }

type UnreadCountHandler struct{ Service *application.Service }

type UnreadCountRequest struct{}

type UnreadCountResponse struct {
	Count int `json:"count"`
}

func (h *UnreadCountHandler) Serve(ctx context.Context, _ *UnreadCountRequest) (UnreadCountResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UnreadCountResponse{}, errorModels.PermissionDeniedServerError()
	}
	count, err := h.Service.UnreadCount(ctx, device.User.BusinessID)
	if err != nil {
		return UnreadCountResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get unread count"}
	}
	return UnreadCountResponse{Count: count}, nil
}
