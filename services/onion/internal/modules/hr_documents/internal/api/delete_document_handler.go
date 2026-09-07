package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DeleteDocumentRoutes() []string { return []string{"/hr/documents/{id}"} }

type DeleteDocumentHandler struct{ Service *application.Service }

type DeleteDocumentRequest struct{}

type DeleteDocumentResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteDocumentHandler) Serve(ctx context.Context, _ *DeleteDocumentRequest) (DeleteDocumentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteDocumentResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteDocumentResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	docID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteDocumentResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid document id")
	}

	err := h.Service.DeleteDocument(ctx, businessID, docID)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return DeleteDocumentResponse{}, onekiterrors.New(http.StatusNotFound, "Document not found")
		}
		logger.Ctx(ctx).Error("failed to delete document", logger.Any("owner_id", device.User.UserID), logger.Any("document_id", docID), logger.Error(err))
		return DeleteDocumentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete document")
	}

	return DeleteDocumentResponse{Success: true}, nil
}
