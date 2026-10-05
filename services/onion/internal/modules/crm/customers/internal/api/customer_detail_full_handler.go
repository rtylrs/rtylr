package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
)

func CustomerDetailFullRoutes() []string {
	return []string{"/crm/customers/detail/full"}
}

type GetCustomerDetailFullHandler struct {
	CRMService *application.Service
}

type GetCustomerDetailFullRequest struct {
	ID int64 `json:"id"`
}

type GetCustomerDetailFullResponse struct {
	Customer            *crmcustomerv1.Customer            `json:"customer"`
	LoyaltyTransactions []*crmloyaltyv1.LoyaltyTransaction `json:"loyalty_transactions"`
	RecentVisits        []TimelineEntryPayload             `json:"recent_visits"`
}

func (h *GetCustomerDetailFullHandler) Validate(_ context.Context, req *GetCustomerDetailFullRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	return nil
}

func (h *GetCustomerDetailFullHandler) Serve(ctx context.Context, req *GetCustomerDetailFullRequest) (GetCustomerDetailFullResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCustomerDetailFullResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCustomerDetailFullResponse{}, onekiterrors.PermissionDenied()
	}

	detail, err := h.CRMService.GetCustomerDetail(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return GetCustomerDetailFullResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return GetCustomerDetailFullResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get customer detail")
	}

	txPayloads := make([]*crmloyaltyv1.LoyaltyTransaction, len(detail.LoyaltyTransactions))
	for i, tx := range detail.LoyaltyTransactions {
		txPayloads[i] = newLoyaltyTransactionPayload(tx)
	}

	visitPayloads := make([]TimelineEntryPayload, len(detail.RecentVisits))
	for i, e := range detail.RecentVisits {
		visitPayloads[i] = TimelineEntryPayload{
			ID:          e.ID,
			CustomerID:  e.CustomerID,
			Type:        e.Type,
			Description: e.Description,
			Metadata:    json.RawMessage(e.Metadata),
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return GetCustomerDetailFullResponse{
		Customer:            newCustomerPayload(detail.Customer),
		LoyaltyTransactions: txPayloads,
		RecentVisits:        visitPayloads,
	}, nil
}
