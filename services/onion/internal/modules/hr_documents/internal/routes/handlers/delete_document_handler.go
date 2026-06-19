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
)

func DeleteDocumentRoutes() []string { return []string{"/hr/documents/{id}"} }

type DeleteDocumentHandler struct{ Service *application.Service }

type DeleteDocumentRequest struct{}

type DeleteDocumentResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteDocumentHandler) Serve(ctx context.Context, _ *DeleteDocumentRequest) (DeleteDocumentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteDocumentResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteDocumentResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	docID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid document id"}
	}

	err := h.Service.DeleteDocument(ctx, businessID, docID)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return DeleteDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Document not found"}
		}
		logger.Ctx(ctx).Error("failed to delete document", logger.Any("owner_id", device.User.UserID), logger.Any("document_id", docID), logger.Error(err))
		return DeleteDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete document"}
	}

	return DeleteDocumentResponse{Success: true}, nil
}
