package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func ListDocumentsRoutes() []string { return []string{"/hr/documents"} }

type ListDocumentsHandler struct{ Service *application.Service }

type ListDocumentsRequest struct {
	EmployeeID int64 `json:"employee_id"`
	Limit      int32 `json:"limit"`
	Offset     int32 `json:"offset"`
}

type ListDocumentsResponse struct {
	Documents []*hrdocumentv1.EmployeeDocumentPayload `json:"documents"`
	Total     int64                                   `json:"total"`
	Limit     int32                                   `json:"limit"`
	Offset    int32                                   `json:"offset"`
}

func (h *ListDocumentsHandler) Serve(ctx context.Context, req *ListDocumentsRequest) (ListDocumentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListDocumentsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListDocumentsResponse{}, errorModels.PermissionDeniedServerError()
	}

	if req.EmployeeID == 0 {
		return ListDocumentsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}

	docs, total, err := h.Service.ListDocuments(ctx, businessID, req.EmployeeID, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list documents", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ListDocumentsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list documents"}
	}

	payloads := make([]*hrdocumentv1.EmployeeDocumentPayload, 0, len(docs))
	for _, d := range docs {
		payloads = append(payloads, documentToProto(d))
	}

	return ListDocumentsResponse{Documents: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
