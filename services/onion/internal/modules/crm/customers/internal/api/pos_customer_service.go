package api

import (
	"context"
	"net/http"
	"strings"

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

func (a *POSCustomerAPI) CreateCustomer(ctx context.Context, req *poscustomerv1.CreateCustomerRequest) (*poscustomerv1.CreateCustomerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	firstName := strings.TrimSpace(req.GetFirstName())
	if firstName == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "First name is required")
	}
	lastName := strings.TrimSpace(req.GetLastName())
	email := strings.TrimSpace(req.GetEmail())
	if email != "" && !strings.Contains(email, "@") {
		return nil, onekiterrors.New(http.StatusBadRequest, "Email address is invalid")
	}
	phone := strings.TrimSpace(req.GetPhone())

	id, err := a.Service.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		Phone:      phone,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create customer")
	}

	zero := int32(0)
	return &poscustomerv1.CreateCustomerResponse{
		Customer: &poscustomerv1.Customer{
			Id:            id,
			FirstName:     firstName,
			LastName:      lastName,
			Email:         email,
			Phone:         phone,
			LoyaltyPoints: &zero,
		},
	}, nil
}

func salesContactToProto(customer domain.SalesContact) *poscustomerv1.Customer {
	points := customer.LoyaltyPoints
	return &poscustomerv1.Customer{
		Id:            customer.ID,
		FirstName:     customer.FirstName,
		LastName:      customer.LastName,
		Email:         customer.Email,
		Phone:         customer.Phone,
		LoyaltyPoints: &points,
	}
}
