package handlers

import (
	"context"
	"net/http"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func CustomerCreateRoutes() []string {
	return []string{"/crm/customers"}
}

type CreateCustomerHandler struct {
	CRMService *application.Service
}

type CreateCustomerRequest struct {
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DateOfBirth string   `json:"dateOfBirth"`
	Gender      string   `json:"gender"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Country     string   `json:"country"`
	Tags        []string `json:"tags"`
	Notes       string   `json:"notes"`
}

type CreateCustomerResponse struct {
	CustomerID int64  `json:"customerId"`
	Message    string `json:"message"`
}

func (h *CreateCustomerHandler) Validate(_ context.Context, req *CreateCustomerRequest) *errorModels.ServerError {
	if req.FirstName == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "First name is required",
		}
	}

	return nil
}

func (h *CreateCustomerHandler) Serve(ctx context.Context, req *CreateCustomerRequest) (CreateCustomerResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.CustomerInput{
		OwnerID:   device.User.UserID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Gender:    req.Gender,
		Address:   req.Address,
		City:      req.City,
		Country:   req.Country,
		Tags:      req.Tags,
		Notes:     req.Notes,
	}

	if req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err == nil {
			input.DateOfBirth = &t
		}
	}

	id, err := h.CRMService.CreateCustomer(ctx, input)
	if err != nil {
		return CreateCustomerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create customer",
		}
	}

	return CreateCustomerResponse{
		CustomerID: id,
		Message:    "Customer created successfully",
	}, nil
}
