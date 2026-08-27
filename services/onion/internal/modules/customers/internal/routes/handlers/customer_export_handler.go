package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
)

func CustomerExportRoutes() []string {
	return []string{"/crm/customers/export"}
}

type ExportCustomerHandler struct {
	CRMService *application.Service
}

type ExportCustomerRequest struct {
	CustomerID int64 `json:"customer_id"`
}

type ExportCustomerResponse struct {
	Customer     *crmcustomerv1.Customer            `json:"customer"`
	Transactions []*crmloyaltyv1.LoyaltyTransaction `json:"transactions"`
	Timeline     []TimelineEntryPayload             `json:"timeline"`
}

func (h *ExportCustomerHandler) Validate(_ context.Context, req *ExportCustomerRequest) error {
	if req.CustomerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	return nil
}

func (h *ExportCustomerHandler) Serve(ctx context.Context, req *ExportCustomerRequest) (ExportCustomerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ExportCustomerResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:export"); err != nil {
		return ExportCustomerResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ExportCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	export, err := h.CRMService.ExportCustomerData(ctx, businessID, req.CustomerID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return ExportCustomerResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return ExportCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export customer data")
	}

	txPayloads := make([]*crmloyaltyv1.LoyaltyTransaction, len(export.Transactions))
	for i, tx := range export.Transactions {
		txPayloads[i] = newLoyaltyTransactionPayload(tx)
	}

	tlPayloads := make([]TimelineEntryPayload, len(export.Timeline))
	for i, e := range export.Timeline {
		tlPayloads[i] = TimelineEntryPayload{
			ID:          e.ID,
			CustomerID:  e.CustomerID,
			Type:        e.Type,
			Description: e.Description,
			Metadata:    json.RawMessage(e.Metadata),
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return ExportCustomerResponse{
		Customer:     newCustomerPayload(export.Customer),
		Transactions: txPayloads,
		Timeline:     tlPayloads,
	}, nil
}
