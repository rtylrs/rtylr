package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
)

type POSCustomerAPI struct {
	Service *application.Service
}

func NewPOSCustomerAPI(service *application.Service) *POSCustomerAPI {
	return &POSCustomerAPI{Service: service}
}

func (a *POSCustomerAPI) SearchCustomers(ctx context.Context, req *poscustomerv1.SearchCustomersRequest) (*poscustomerv1.SearchCustomersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	customers, err := a.Service.SearchSalesContacts(ctx, businessID, req.GetQuery())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to search customers")
	}

	payloads := make([]*poscustomerv1.Customer, 0, len(customers))
	for _, customer := range customers {
		payloads = append(payloads, salesContactToProto(customer))
	}

	return &poscustomerv1.SearchCustomersResponse{Customers: payloads}, nil
}

func salesContactToProto(customer domain.SalesContact) *poscustomerv1.Customer {
	return &poscustomerv1.Customer{
		Id:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
	}
}
