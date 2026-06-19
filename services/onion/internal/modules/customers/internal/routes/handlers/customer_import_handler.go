package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CustomerImportRoutes() []string {
	return []string{"/crm/customers/import"}
}

type ImportCustomersHandler struct {
	CRMService *application.Service
}

type ImportCustomerItem struct {
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DateOfBirth string   `json:"date_of_birth"`
	Gender      string   `json:"gender"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Country     string   `json:"country"`
	Tags        []string `json:"tags"`
	Notes       string   `json:"notes"`
}

type ImportCustomersRequest struct {
	Customers []ImportCustomerItem `json:"customers"`
}

type ImportCustomersResponse struct {
	TotalProcessed int    `json:"total_processed"`
	Imported       int    `json:"imported"`
	Failed         int    `json:"failed"`
	Message        string `json:"message"`
}

func (h *ImportCustomersHandler) Validate(_ context.Context, req *ImportCustomersRequest) *errorModels.ServerError {
	if len(req.Customers) == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "At least one customer is required",
		}
	}

	return nil
}

func (h *ImportCustomersHandler) Serve(ctx context.Context, req *ImportCustomersRequest) (ImportCustomersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ImportCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ImportCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}

	items := make([]domain.CustomerImportItem, len(req.Customers))
	for i, c := range req.Customers {
		items[i] = domain.CustomerImportItem{
			FirstName:   c.FirstName,
			LastName:    c.LastName,
			Email:       c.Email,
			Phone:       c.Phone,
			DateOfBirth: c.DateOfBirth,
			Gender:      c.Gender,
			Address:     c.Address,
			City:        c.City,
			Country:     c.Country,
			Tags:        c.Tags,
			Notes:       c.Notes,
		}
	}

	result, err := h.CRMService.ImportCustomers(ctx, businessID, items)
	if err != nil {
		return ImportCustomersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to import customers",
		}
	}

	return ImportCustomersResponse{
		TotalProcessed: result.TotalProcessed,
		Imported:       result.Imported,
		Failed:         result.Failed,
		Message:        "Import completed",
	}, nil
}
