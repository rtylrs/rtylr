package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func UploadDocumentRoutes() []string { return []string{"/hr/documents"} }

type UploadDocumentHandler struct{ Service *application.Service }

type UploadDocumentRequest struct {
	EmployeeID   int64   `json:"employee_id"`
	DocumentType string  `json:"document_type"`
	Title        string  `json:"title"`
	StorageMode  string  `json:"storage_mode"`
	FileURL      string  `json:"file_url"`
	ExternalURL  string  `json:"external_url"`
	ExpiryDate   *string `json:"expiry_date,omitempty"`
}

type UploadDocumentResponse struct {
	Document *hrdocumentv1.EmployeeDocumentPayload `json:"document"`
}

func (h *UploadDocumentHandler) Validate(_ context.Context, req *UploadDocumentRequest) error {
	if req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "title is required")
	}
	if strings.TrimSpace(req.StorageMode) == "" {
		return onekiterrors.New(http.StatusBadRequest, "storage_mode is required")
	}
	if req.ExpiryDate != nil {
		if _, err := parseExpiryDate(*req.ExpiryDate); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "expiry_date must be a valid date")
		}
	}
	return nil
}

func (h *UploadDocumentHandler) Serve(ctx context.Context, req *UploadDocumentRequest) (UploadDocumentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UploadDocumentResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireDocumentPermission(ctx, "hr:employees:manage"); err != nil {
		return UploadDocumentResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UploadDocumentResponse{}, onekiterrors.PermissionDenied()
	}

	fileURL := strings.TrimSpace(req.FileURL)
	externalURL := strings.TrimSpace(req.ExternalURL)
	if req.StorageMode == "linked" {
		if externalURL == "" {
			externalURL = fileURL
		}
		fileURL = ""
	}
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		parsed, parseErr := parseExpiryDate(*req.ExpiryDate)
		if parseErr != nil {
			return UploadDocumentResponse{}, onekiterrors.New(http.StatusBadRequest, "expiry_date must be a valid date")
		}
		expiryDate = &parsed
	}

	doc, err := h.Service.UploadDocument(ctx, businessID, domain.CreateDocumentParams{
		EmployeeID:   req.EmployeeID,
		DocumentType: req.DocumentType,
		Title:        req.Title,
		StorageMode:  req.StorageMode,
		FileURL:      fileURL,
		ExternalURL:  externalURL,
		ExpiryDate:   expiryDate,
		UploadedBy:   device.User.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrStorageModeInvalid):
			return UploadDocumentResponse{}, onekiterrors.New(http.StatusBadRequest, "storage_mode must be 'stored' or 'linked'; stored requires file_url, linked requires external_url")
		case errors.Is(err, domain.ErrDocumentNotFound):
			return UploadDocumentResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found")
		default:
			logger.Ctx(ctx).Error("failed to upload document", logger.Any("owner_id", device.User.UserID), logger.Error(err))
			return UploadDocumentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to upload document")
		}
	}

	return UploadDocumentResponse{Document: documentToProto(*doc)}, nil
}

func parseExpiryDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}
