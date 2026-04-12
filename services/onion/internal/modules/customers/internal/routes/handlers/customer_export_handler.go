package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
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
	Customer     CustomerPayload             `json:"customer"`
	Transactions []LoyaltyTransactionPayload `json:"transactions"`
	Timeline     []TimelineEntryPayload      `json:"timeline"`
}

func (h *ExportCustomerHandler) Validate(_ context.Context, req *ExportCustomerRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	return nil
}

func (h *ExportCustomerHandler) Serve(ctx context.Context, req *ExportCustomerRequest) (ExportCustomerResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ExportCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}

	export, err := h.CRMService.ExportCustomerData(ctx, device.User.UserID, req.CustomerID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return ExportCustomerResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Customer not found",
			}
		}

		return ExportCustomerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to export customer data",
		}
	}

	txPayloads := make([]LoyaltyTransactionPayload, len(export.Transactions))
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
