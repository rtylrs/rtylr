package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/application"
	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

var expiryDateLayouts = []string{
	"2006-01-02",
	time.RFC3339,
	"2006-01-02 15:04:05",
}

// listDocumentsPageSize bounds ListDocuments server-side. The voxie
// hr.document.v1.ListDocumentsRequest contract has no pagination fields yet,
// so the endpoint returns at most one maximum-size page instead of every row.
const listDocumentsPageSize int32 = 100

type DocumentAPI struct {
	service *application.Service
}

func NewDocumentAPI(service *application.Service) *DocumentAPI {
	return &DocumentAPI{service: service}
}

func (a *DocumentAPI) ListDocuments(ctx context.Context, _ *hrdocumentv1.ListDocumentsRequest) (*hrdocumentv1.ListDocumentsResponse, error) {
	if err := requireDocumentPermission(ctx, "hr:employees:view"); err != nil {
		return nil, err
	}

	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	items, err := a.service.ListPaginated(ctx, businessID, listDocumentsPageSize, 0)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list documents")
	}

	documents := make([]*hrdocumentv1.Document, 0, len(items))
	for _, item := range items {
		documents = append(documents, newDocument(item))
	}

	return &hrdocumentv1.ListDocumentsResponse{Documents: documents}, nil
}

func (a *DocumentAPI) CreateDocument(ctx context.Context, req *hrdocumentv1.CreateDocumentRequest) (*hrdocumentv1.CreateDocumentResponse, error) {
	if err := requireDocumentPermission(ctx, "hr:employees:manage"); err != nil {
		return nil, err
	}
	if req == nil || req.GetName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if req.GetStaffId() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "staff_id must be positive or zero")
	}
	if req.GetExpiryDate() != "" && !isValidExpiryDate(req.GetExpiryDate()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "expiry_date is not a valid date")
	}

	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	id, err := a.service.Create(ctx, businessID, domain.DocumentInput{
		StaffID:    req.GetStaffId(),
		Name:       req.GetName(),
		Type:       req.GetType(),
		FileURL:    req.GetFileUrl(),
		ExpiryDate: req.GetExpiryDate(),
	})
	if errors.Is(err, domain.ErrDocumentStaffNotFound) {
		return nil, onekiterrors.New(http.StatusNotFound, "Staff member not found in this business")
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create document")
	}

	return &hrdocumentv1.CreateDocumentResponse{Id: id, Message: "Document created successfully"}, nil
}

func (a *DocumentAPI) DeleteDocument(ctx context.Context, req *hrdocumentv1.DeleteDocumentRequest) (*hrdocumentv1.DeleteDocumentResponse, error) {
	if err := requireDocumentPermission(ctx, "hr:employees:manage"); err != nil {
		return nil, err
	}
	if req == nil || req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "id is required")
	}

	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrDocumentNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Document not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete document")
	}

	return &hrdocumentv1.DeleteDocumentResponse{Message: "Document deleted successfully"}, nil
}

func requireDocumentPermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil ||
		!authPkg.HasAppPermission(device.User, authorization.AppHR, permission) ||
		!authPkg.HasAnyAppScope(device.User, authorization.AppHR) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func resolveBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func newDocument(item domain.Document) *hrdocumentv1.Document {
	return &hrdocumentv1.Document{
		Id:         item.ID,
		OwnerId:    item.OwnerID,
		StaffId:    item.StaffID,
		Name:       item.Name,
		Type:       item.Type,
		FileUrl:    item.FileURL,
		ExpiryDate: formatTime(item.ExpiryDate),
		CreatedAt:  formatTime(item.CreatedAt),
	}
}

func isValidExpiryDate(value string) bool {
	for _, layout := range expiryDateLayouts {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}
	return false
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
