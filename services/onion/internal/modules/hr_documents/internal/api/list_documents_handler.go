package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

type LegacyDocumentResponse struct {
	ID         int64  `json:"id"`
	StaffID    int64  `json:"staff_id"`
	StaffName  string `json:"staff_name"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	UploadedAt string `json:"uploaded_at"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	SizeKB     int64  `json:"size_kb"`
	URL        string `json:"url,omitempty"`
}

func (h *ListDocumentsHandler) Serve(ctx context.Context, req *ListDocumentsRequest) (any, error) {
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

	if req.EmployeeID < 0 {
		return ListDocumentsResponse{}, onekiterrors.New(http.StatusBadRequest, "employee_id must be positive or zero")
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

	docs, total, err := h.Service.ListDocuments(ctx, businessID, req.EmployeeID, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list documents", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return ListDocumentsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list documents")
	}

	if req.EmployeeID == 0 {
		return legacyDocumentsResponse(docs), nil
	}

	payloads := make([]*hrdocumentv1.EmployeeDocumentPayload, 0, len(docs))
	for _, d := range docs {
		payloads = append(payloads, documentToProto(d))
	}

	return ListDocumentsResponse{Documents: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}

func legacyDocumentsResponse(records []domain.EmployeeDocumentRecord) []LegacyDocumentResponse {
	documents := make([]LegacyDocumentResponse, 0, len(records))
	for _, record := range records {
		uploadedAt := ""
		if !record.CreatedAt.IsZero() {
			uploadedAt = record.CreatedAt.UTC().Format(time.RFC3339)
		}
		expiresAt := ""
		if record.ExpiryDate != nil {
			expiresAt = record.ExpiryDate.UTC().Format("2006-01-02")
		}
		fileURL := record.FileURL
		if record.ExternalURL != "" {
			fileURL = record.ExternalURL
		}
		documents = append(documents, LegacyDocumentResponse{
			ID:         record.ID,
			StaffID:    record.EmployeeID,
			StaffName:  record.EmployeeName,
			Name:       record.Title,
			Type:       legacyDocumentType(record.DocumentType),
			UploadedAt: uploadedAt,
			ExpiresAt:  expiresAt,
			SizeKB:     0,
			URL:        fileURL,
		})
	}
	return documents
}

func legacyDocumentType(documentType string) string {
	switch strings.ToLower(strings.TrimSpace(documentType)) {
	case "contract":
		return "Contract"
	case "tax_form", "w4":
		return "W4"
	case "id_card", "i-9":
		return "I-9"
	case "nda":
		return "NDA"
	case "offer_letter", "offer letter":
		return "Offer Letter"
	default:
		return "Other"
	}
}
