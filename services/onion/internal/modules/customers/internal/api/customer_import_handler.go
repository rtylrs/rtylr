package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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
	Dedupe    bool                 `json:"dedupe"`
}

type ImportCustomersResponse struct {
	TotalProcessed int    `json:"total_processed"`
	Imported       int    `json:"imported"`
	Failed         int    `json:"failed"`
	Message        string `json:"message"`
}

// maxImportBatchSize bounds one import request. Imports run sequential
// inserts; without a cap a single request could push an unbounded number of
// PII rows in one shot.
const maxImportBatchSize = 1000

func (h *ImportCustomersHandler) Validate(_ context.Context, req *ImportCustomersRequest) error {
	if len(req.Customers) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "At least one customer is required")
	}
	if len(req.Customers) > maxImportBatchSize {
		return onekiterrors.New(http.StatusBadRequest, "Too many customers per import batch")
	}

	return nil
}

func (h *ImportCustomersHandler) Serve(ctx context.Context, req *ImportCustomersRequest) (ImportCustomersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ImportCustomersResponse{}, onekiterrors.PermissionDenied()
	}
	// Bulk-inserting PII is destructive/policy-changing, same as export,
	// delete and merge.
	if err := requireCRMFeaturePermission(ctx, "crm:customers:create"); err != nil {
		return ImportCustomersResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ImportCustomersResponse{}, onekiterrors.PermissionDenied()
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

	result, err := h.CRMService.ImportCustomers(ctx, businessID, device.User.UserID, items, req.Dedupe)
	if err != nil {
		return ImportCustomersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to import customers")
	}

	return ImportCustomersResponse{
		TotalProcessed: result.TotalProcessed,
		Imported:       result.Imported,
		Failed:         result.Failed,
		Message:        "Import completed",
	}, nil
}
