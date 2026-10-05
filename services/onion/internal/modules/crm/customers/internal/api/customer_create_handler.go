package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CustomerCreateRoutes() []string {
	return []string{"/crm/customers"}
}

type CreateCustomerHandler struct {
	CRMService *application.Service
}

type CreateCustomerRequest struct {
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	DateOfBirth string          `json:"date_of_birth"`
	Gender      string          `json:"gender"`
	Address     string          `json:"address"`
	City        string          `json:"city"`
	Country     string          `json:"country"`
	Tags        []string        `json:"tags"`
	Notes       string          `json:"notes"`
	Consent     map[string]bool `json:"consent"`
}

type CreateCustomerResponse struct {
	CustomerID int64  `json:"customer_id"`
	Message    string `json:"message"`
}

func (h *CreateCustomerHandler) Validate(_ context.Context, req *CreateCustomerRequest) error {
	if req.FirstName == "" {
		return onekiterrors.New(http.StatusBadRequest, "First name is required")
	}

	return nil
}

func (h *CreateCustomerHandler) Serve(ctx context.Context, req *CreateCustomerRequest) (CreateCustomerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.CustomerInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Gender:     req.Gender,
		Address:    req.Address,
		City:       req.City,
		Country:    req.Country,
		Tags:       req.Tags,
		Notes:      req.Notes,
	}
	if req.Consent != nil {
		consent, err := json.Marshal(req.Consent)
		if err != nil {
			return CreateCustomerResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid consent data")
		}
		input.Consent = consent
	}

	if req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return CreateCustomerResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid date of birth format, expected YYYY-MM-DD")
		}
		input.DateOfBirth = &t
	}

	id, err := h.CRMService.CreateCustomer(ctx, input)
	if err != nil {
		return CreateCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create customer")
	}

	return CreateCustomerResponse{
		CustomerID: id,
		Message:    "Customer created successfully",
	}, nil
}
