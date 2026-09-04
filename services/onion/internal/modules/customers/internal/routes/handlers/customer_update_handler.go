package handlers

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
	ID          int64           `json:"id"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	DateOfBirth string          `json:"date_of_birth"`
	Gender      string          `json:"gender"`
	Address     string          `json:"address"`
	City        string          `json:"city"`
	Country     string          `json:"country"`
	LoyaltyTier string          `json:"loyalty_tier"`
	Tags        []string        `json:"tags"`
	Notes       string          `json:"notes"`
	Consent     map[string]bool `json:"consent"`
	IsActive    *bool           `json:"is_active"`
}

type UpdateCustomerResponse struct {
	Message string `json:"message"`
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

	input := domain.CustomerUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		Gender:      req.Gender,
		Address:     req.Address,
		City:        req.City,
		Country:     req.Country,
		LoyaltyTier: req.LoyaltyTier,
		Tags:        req.Tags,
		Notes:       req.Notes,
		IsActive:    req.IsActive,
	}
	if req.Consent != nil {
		consent, err := json.Marshal(req.Consent)
		if err != nil {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid consent data")
		}
		input.Consent = consent
	}

	if req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid date of birth format, expected YYYY-MM-DD")
		}
		input.DateOfBirth = &t
	}

	if err := h.CRMService.UpdateCustomer(ctx, input); err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return UpdateCustomerResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return UpdateCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update customer")
	}

	return UpdateCustomerResponse{Message: "Customer updated successfully"}, nil
}
