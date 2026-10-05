package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListExpiringDocumentsRoutes() []string { return []string{"/hr/documents/expiring"} }

type ListExpiringDocumentsHandler struct{ Service *application.Service }

type ListExpiringDocumentsRequest struct {
	Days   int32 `json:"days"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (h *ListExpiringDocumentsHandler) Serve(ctx context.Context, req *ListExpiringDocumentsRequest) (any, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListDocumentsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireDocumentPermission(ctx, "hr:employees:view"); err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListDocumentsResponse{}, onekiterrors.PermissionDenied()
	}

	days := req.Days
	if days == 0 {
		days = 30
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	docs, _, err := h.Service.ListExpiringDocuments(ctx, businessID, days, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list expiring documents", logger.Any("owner_id", device.User.UserID), logger.Any("days", days), logger.Error(err))
		return ListDocumentsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list expiring documents")
	}

	return legacyDocumentsResponse(docs), nil
}
