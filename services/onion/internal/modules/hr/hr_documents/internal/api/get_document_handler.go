package api

import (
	"context"
	"errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func GetDocumentRoutes() []string { return []string{"/hr/documents/{id}"} }

type GetDocumentHandler struct{ Service *application.Service }

type GetDocumentRequest struct{}

type GetDocumentResponse struct {
	Document *hrdocumentv1.EmployeeDocumentPayload `json:"document"`
}

func (h *GetDocumentHandler) Serve(ctx context.Context, _ *GetDocumentRequest) (GetDocumentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetDocumentResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireDocumentPermission(ctx, "hr:employees:view"); err != nil {
		return GetDocumentResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetDocumentResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := serverHttp.PathParam(ctx, "id")
	docID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetDocumentResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid document id")
	}

	doc, err := h.Service.GetDocument(ctx, businessID, docID)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return GetDocumentResponse{}, onekiterrors.New(http.StatusNotFound, "Document not found")
		}
		logger.Ctx(ctx).Error("failed to get document", logger.Any("owner_id", device.User.UserID), logger.Any("document_id", docID), logger.Error(err))
		return GetDocumentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get document")
	}

	return GetDocumentResponse{Document: documentToProto(*doc)}, nil
}
