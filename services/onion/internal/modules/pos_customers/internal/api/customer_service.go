package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/application"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
)

type CustomerAPI struct {
	Service *application.Service
}

func NewCustomerAPI(service *application.Service) *CustomerAPI {
	return &CustomerAPI{Service: service}
}

func (a *CustomerAPI) SearchCustomers(ctx context.Context, req *poscustomerv1.SearchCustomersRequest) (*poscustomerv1.SearchCustomersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	customers, err := a.Service.Search(ctx, businessID, req.GetQuery())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to search customers")
	}

	payloads := make([]*poscustomerv1.Customer, 0, len(customers))
	for _, customer := range customers {
		payloads = append(payloads, customerToProto(customer))
	}

	return &poscustomerv1.SearchCustomersResponse{Customers: payloads}, nil
}

func customerToProto(customer domain.Customer) *poscustomerv1.Customer {
	return &poscustomerv1.Customer{
		Id:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
	}
}
