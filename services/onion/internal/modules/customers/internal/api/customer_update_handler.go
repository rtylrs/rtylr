package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CustomerUpdateRoutes() []string {
	return []string{"/crm/customers"}
}

type UpdateCustomerHandler struct {
	CRMService *application.Service
}

type UpdateCustomerRequest struct {
	ID          int64            `json:"id"`
	FirstName   string           `json:"first_name"`
	LastName    *string          `json:"last_name"`
	Email       *string          `json:"email"`
	Phone       *string          `json:"phone"`
	DateOfBirth *string          `json:"date_of_birth"`
	Gender      *string          `json:"gender"`
	Address     *string          `json:"address"`
	City        *string          `json:"city"`
	Country     *string          `json:"country"`
	LoyaltyTier *string          `json:"loyalty_tier"`
	Tags        *[]string        `json:"tags"`
	Notes       *string          `json:"notes"`
	Consent     *map[string]bool `json:"consent"`
	IsActive    *bool            `json:"is_active"`
}

type UpdateCustomerResponse struct {
	Message string `json:"message"`
}

func customerUpdateValue[T any](value *T, fallback T) T {
	if value != nil {
		return *value
	}
	return fallback
}

func (h *UpdateCustomerHandler) Validate(_ context.Context, req *UpdateCustomerRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	if req.FirstName == "" {
		return onekiterrors.New(http.StatusBadRequest, "First name is required")
	}

	return nil
}

func (h *UpdateCustomerHandler) Serve(ctx context.Context, req *UpdateCustomerRequest) (UpdateCustomerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	current, err := h.CRMService.GetCustomer(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}
		return UpdateCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update customer")
	}

	dateOfBirth := current.DateOfBirth
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, parseErr := time.Parse("2006-01-02", *req.DateOfBirth)
		if parseErr != nil {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid date of birth format, expected YYYY-MM-DD")
		}
		dateOfBirth = &parsed
	}

	consent := current.Consent
	if req.Consent != nil {
		consent, err = json.Marshal(*req.Consent)
		if err != nil {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid consent data")
		}
	}

	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	input := domain.CustomerUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		FirstName:   req.FirstName,
		LastName:    customerUpdateValue(req.LastName, current.LastName),
		Email:       customerUpdateValue(req.Email, current.Email),
		Phone:       customerUpdateValue(req.Phone, current.Phone),
		DateOfBirth: dateOfBirth,
		Gender:      customerUpdateValue(req.Gender, current.Gender),
		Address:     customerUpdateValue(req.Address, current.Address),
		City:        customerUpdateValue(req.City, current.City),
		Country:     customerUpdateValue(req.Country, current.Country),
		LoyaltyTier: customerUpdateValue(req.LoyaltyTier, current.LoyaltyTier),
		Tags:        customerUpdateValue(req.Tags, current.Tags),
		Notes:       customerUpdateValue(req.Notes, current.Notes),
		Consent:     consent,
		IsActive:    &isActive,
	}

	if err := h.CRMService.UpdateCustomer(ctx, input); err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return UpdateCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update customer")
	}

	return UpdateCustomerResponse{Message: "Customer updated successfully"}, nil
}
