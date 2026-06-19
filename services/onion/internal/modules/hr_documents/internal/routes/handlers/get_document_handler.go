package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func GetDocumentRoutes() []string { return []string{"/hr/documents/{id}"} }

type GetDocumentHandler struct{ Service *application.Service }

type GetDocumentRequest struct{}

type GetDocumentResponse struct {
	Document *hrdocumentv1.EmployeeDocumentPayload `json:"document"`
}

func (h *GetDocumentHandler) Serve(ctx context.Context, _ *GetDocumentRequest) (GetDocumentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetDocumentResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetDocumentResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	docID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid document id"}
	}

	doc, err := h.Service.GetDocument(ctx, businessID, docID)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return GetDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Document not found"}
		}
		logger.Ctx(ctx).Error("failed to get document", logger.Any("owner_id", device.User.UserID), logger.Any("document_id", docID), logger.Error(err))
		return GetDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get document"}
	}

	return GetDocumentResponse{Document: documentToProto(*doc)}, nil
}
