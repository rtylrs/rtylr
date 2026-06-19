package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func UploadDocumentRoutes() []string { return []string{"/hr/documents"} }

type UploadDocumentHandler struct{ Service *application.Service }

type UploadDocumentRequest struct {
	EmployeeID   int64  `json:"employee_id"`
	DocumentType string `json:"document_type"`
	Title        string `json:"title"`
	StorageMode  string `json:"storage_mode"`
	FileURL      string `json:"file_url"`
	ExternalURL  string `json:"external_url"`
}

type UploadDocumentResponse struct {
	Document *hrdocumentv1.EmployeeDocumentPayload `json:"document"`
}

func (h *UploadDocumentHandler) Validate(_ context.Context, req *UploadDocumentRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if strings.TrimSpace(req.Title) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "title is required"}
	}
	if strings.TrimSpace(req.StorageMode) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "storage_mode is required"}
	}
	return nil
}

func (h *UploadDocumentHandler) Serve(ctx context.Context, req *UploadDocumentRequest) (UploadDocumentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UploadDocumentResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UploadDocumentResponse{}, errorModels.PermissionDeniedServerError()
	}

	doc, err := h.Service.UploadDocument(ctx, businessID, domain.CreateDocumentParams{
		EmployeeID:   req.EmployeeID,
		DocumentType: req.DocumentType,
		Title:        req.Title,
		StorageMode:  req.StorageMode,
		FileURL:      req.FileURL,
		ExternalURL:  req.ExternalURL,
		UploadedBy:   device.User.UserID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrStorageModeInvalid) {
			return UploadDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "storage_mode must be 'stored' or 'linked'; stored requires file_url, linked requires external_url"}
		}
		logger.Ctx(ctx).Error("failed to upload document", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UploadDocumentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to upload document"}
	}

	return UploadDocumentResponse{Document: documentToProto(*doc)}, nil
}
