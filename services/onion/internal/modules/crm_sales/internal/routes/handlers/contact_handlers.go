package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// GET /crm/sales/contacts
// ---------------------------------------------------------------------------

func ListContactsRoutes() []string { return []string{"/crm/sales/contacts"} }

type ListContactsHandler struct{ Service *application.Service }

type ListContactsRequest struct {
	LeadID *int64 `json:"lead_id"`
}

type ListContactsResponse struct {
	Contacts []domain.Contact `json:"contacts"`
}

func (h *ListContactsHandler) Serve(ctx context.Context, req *ListContactsRequest) (ListContactsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListContactsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListContactsResponse{}, errorModels.PermissionDeniedServerError()
	}

	contacts, err := h.Service.ListContacts(ctx, businessID, req.LeadID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list contacts", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListContactsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list contacts"}
	}

	return ListContactsResponse{Contacts: contacts}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/leads/{id}/touchpoints  (lead_id from URL)
// ---------------------------------------------------------------------------

func ListLeadTouchpointsRoutes() []string { return []string{"/crm/sales/leads/{id}/touchpoints"} }

type (
	ListLeadTouchpointsHandler  struct{ Service *application.Service }
	ListLeadTouchpointsRequest  struct{}
	ListLeadTouchpointsResponse struct {
		Touchpoints []domain.Touchpoint `json:"touchpoints"`
	}
)

func (h *ListLeadTouchpointsHandler) Serve(ctx context.Context, _ *ListLeadTouchpointsRequest) (ListLeadTouchpointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLeadTouchpointsResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	leadID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ListLeadTouchpointsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid lead id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLeadTouchpointsResponse{}, errorModels.PermissionDeniedServerError()
	}

	tps, err := h.Service.ListTouchpoints(ctx, businessID, leadID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list touchpoints", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", leadID), logger.Error(err))
		return ListLeadTouchpointsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list touchpoints"}
	}

	return ListLeadTouchpointsResponse{Touchpoints: tps}, nil
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
	Contact domain.Contact `json:"contact"`
}

func (h *CreateContactHandler) Validate(_ context.Context, req *CreateContactRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.LeadID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "lead_id is required"}
	}
	return nil
}

func (h *CreateContactHandler) Serve(ctx context.Context, req *CreateContactRequest) (CreateContactResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateContactResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateContactResponse{}, errorModels.PermissionDeniedServerError()
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
		logger.Ctx(ctx).Error("failed to create contact", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create contact"}
	}

	return CreateContactResponse{Contact: contact}, nil
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
	Contact domain.Contact `json:"contact"`
}

func (h *UpdateContactHandler) Validate(_ context.Context, req *UpdateContactRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateContactHandler) Serve(ctx context.Context, req *UpdateContactRequest) (UpdateContactResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateContactResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid contact id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateContactResponse{}, errorModels.PermissionDeniedServerError()
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
			return UpdateContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Contact not found"}
		}
		logger.Ctx(ctx).Error("failed to update contact", logger.Any("owner_id", device.User.UserID), logger.Any("contact_id", id), logger.Error(err))
		return UpdateContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update contact"}
	}

	return UpdateContactResponse{Contact: contact}, nil
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

func (h *DeleteContactHandler) Serve(ctx context.Context, _ *DeleteContactRequest) (DeleteContactResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteContactResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid contact id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteContactResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteContact(ctx, id, businessID); err != nil {
		if errors.Is(err, domain.ErrContactNotFound) {
			return DeleteContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Contact not found"}
		}
		logger.Ctx(ctx).Error("failed to delete contact", logger.Any("owner_id", device.User.UserID), logger.Any("contact_id", id), logger.Error(err))
		return DeleteContactResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete contact"}
	}

	return DeleteContactResponse{Success: true}, nil
}
