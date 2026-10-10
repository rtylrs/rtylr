package api

import (
	"context"
	"errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// GET /crm/sales/contacts
// ---------------------------------------------------------------------------

func ListContactsRoutes() []string { return []string{"/crm/sales/contacts"} }

type ListContactsHandler struct{ Service *application.Service }

type ListContactsRequest struct {
	LeadID *int64 `json:"lead_id"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

type ListContactsResponse struct {
	Contacts []contactPayload `json:"contacts"`
	Total    int64            `json:"total"`
	Limit    int32            `json:"limit"`
	Offset   int32            `json:"offset"`
}

func (h *ListContactsHandler) Serve(ctx context.Context, req *ListContactsRequest) (ListContactsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListContactsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return ListContactsResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListContactsResponse{}, onekiterrors.PermissionDenied()
	}

	const maxContactPageSize int32 = 100
	if req.Limit <= 0 || req.Limit > maxContactPageSize {
		req.Limit = maxContactPageSize
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	page, err := h.Service.ListContacts(ctx, businessID, req.LeadID, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list contacts", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListContactsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list contacts")
	}

	return ListContactsResponse{Contacts: payloadContactList(page.Contacts), Total: page.Total, Limit: page.Limit, Offset: page.Offset}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/leads/{id}/touchpoints  (lead_id from URL)
// ---------------------------------------------------------------------------

func ListLeadTouchpointsRoutes() []string { return []string{"/crm/sales/leads/{id}/touchpoints"} }

type (
	ListLeadTouchpointsHandler  struct{ Service *application.Service }
	ListLeadTouchpointsRequest  struct{}
	ListLeadTouchpointsResponse struct {
		Touchpoints []touchpointPayload `json:"touchpoints"`
	}
)

func (h *ListLeadTouchpointsHandler) Serve(ctx context.Context, _ *ListLeadTouchpointsRequest) (ListLeadTouchpointsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLeadTouchpointsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return ListLeadTouchpointsResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	leadID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ListLeadTouchpointsResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid lead id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLeadTouchpointsResponse{}, onekiterrors.PermissionDenied()
	}

	tps, err := h.Service.ListTouchpoints(ctx, businessID, leadID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list touchpoints", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", leadID), logger.Error(err))
		return ListLeadTouchpointsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list touchpoints")
	}

	return ListLeadTouchpointsResponse{Touchpoints: payloadTouchpointList(tps)}, nil
}

// ---------------------------------------------------------------------------
// POST /crm/sales/contacts
// ---------------------------------------------------------------------------

func CreateContactRoutes() []string { return []string{"/crm/sales/contacts"} }

type CreateContactHandler struct{ Service *application.Service }

type CreateContactRequest struct {
	LeadID    int64  `json:"lead_id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	LinkedIn  string `json:"linkedin"`
	IsPrimary bool   `json:"is_primary"`
	Notes     string `json:"notes"`
}

type CreateContactResponse struct {
	Contact contactPayload `json:"contact"`
}

func (h *CreateContactHandler) Validate(_ context.Context, req *CreateContactRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if req.LeadID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "lead_id is required")
	}
	return nil
}

func (h *CreateContactHandler) Serve(ctx context.Context, req *CreateContactRequest) (CreateContactResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateContactResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:create"); err != nil {
		return CreateContactResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateContactResponse{}, onekiterrors.PermissionDenied()
	}

	contact, err := h.Service.CreateContact(ctx, businessID, domain.CreateContactRequest{
		LeadID:    req.LeadID,
		Name:      req.Name,
		Role:      req.Role,
		Email:     req.Email,
		Phone:     req.Phone,
		LinkedIn:  req.LinkedIn,
		IsPrimary: req.IsPrimary,
		Notes:     req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return CreateContactResponse{}, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		logger.Ctx(ctx).Error("failed to create contact", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateContactResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create contact")
	}

	return CreateContactResponse{Contact: payloadContact(contact)}, nil
}

// ---------------------------------------------------------------------------
// PUT /crm/sales/contacts/{id}
// ---------------------------------------------------------------------------

func UpdateContactRoutes() []string { return []string{"/crm/sales/contacts/{id}"} }

type UpdateContactHandler struct{ Service *application.Service }

type UpdateContactRequest struct {
	Name      string `json:"name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	LinkedIn  string `json:"linkedin"`
	IsPrimary bool   `json:"is_primary"`
	Notes     string `json:"notes"`
}

type UpdateContactResponse struct {
	Contact contactPayload `json:"contact"`
}

func (h *UpdateContactHandler) Validate(_ context.Context, req *UpdateContactRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateContactHandler) Serve(ctx context.Context, req *UpdateContactRequest) (UpdateContactResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateContactResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
		return UpdateContactResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateContactResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid contact id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateContactResponse{}, onekiterrors.PermissionDenied()
	}

	contact, err := h.Service.UpdateContact(ctx, id, businessID, domain.UpdateContactRequest{
		Name:      req.Name,
		Role:      req.Role,
		Email:     req.Email,
		Phone:     req.Phone,
		LinkedIn:  req.LinkedIn,
		IsPrimary: req.IsPrimary,
		Notes:     req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			return UpdateContactResponse{}, onekiterrors.New(http.StatusNotFound, "Contact not found")
		}
		logger.Ctx(ctx).Error("failed to update contact", logger.Any("owner_id", device.User.UserID), logger.Any("contact_id", id), logger.Error(err))
		return UpdateContactResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update contact")
	}

	return UpdateContactResponse{Contact: payloadContact(contact)}, nil
}

// ---------------------------------------------------------------------------
// DELETE /crm/sales/contacts/{id}
// ---------------------------------------------------------------------------

func DeleteContactRoutes() []string { return []string{"/crm/sales/contacts/{id}"} }

type (
	DeleteContactHandler  struct{ Service *application.Service }
	DeleteContactRequest  struct{}
	DeleteContactResponse struct {
		Success bool `json:"success"`
	}
)

func (h *DeleteContactHandler) Serve(ctx context.Context, _ *DeleteContactRequest) (DeleteContactResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteContactResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:delete"); err != nil {
		return DeleteContactResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteContactResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid contact id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteContactResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.DeleteContact(ctx, id, businessID); err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			return DeleteContactResponse{}, onekiterrors.New(http.StatusNotFound, "Contact not found")
		}
		logger.Ctx(ctx).Error("failed to delete contact", logger.Any("owner_id", device.User.UserID), logger.Any("contact_id", id), logger.Error(err))
		return DeleteContactResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete contact")
	}

	return DeleteContactResponse{Success: true}, nil
}
