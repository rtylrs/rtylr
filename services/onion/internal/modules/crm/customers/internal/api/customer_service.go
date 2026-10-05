package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	customerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
	dashboardv1 "github.com/voxire/voxie/gen/go/crm/dashboard/v1"
)

// Reuse the legacy handlers so validation and tenant resolution stay consistent.
type CustomerAPI struct{ service *application.Service }

func NewCustomerAPI(service *application.Service) *CustomerAPI { return &CustomerAPI{service: service} }

func (a *CustomerAPI) ListCustomers(ctx context.Context, req *customerv1.ListCustomersRequest) (*customerv1.ListCustomersResponse, error) {
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return nil, err
	}
	handler := &ListCustomersHandler{CRMService: a.service}
	input := &ListCustomersRequest{Search: req.GetSearch(), Tier: req.GetTier(), Tag: req.GetTag(), Page: int(req.GetPage()), PageSize: int(req.GetPageSize())}
	result, err := handler.Serve(ctx, input)
	if err != nil {
		return nil, err
	}
	return &customerv1.ListCustomersResponse{Customers: result.Customers, Total: int64(result.Total), Page: int64(result.Page), PageSize: int64(result.PageSize)}, nil
}
func (a *CustomerAPI) CreateCustomer(ctx context.Context, req *customerv1.CreateCustomerRequest) (*customerv1.CreateCustomerResponse, error) {
	if err := requireCRMFeaturePermission(ctx, "crm:customers:create"); err != nil {
		return nil, err
	}
	handler := &CreateCustomerHandler{CRMService: a.service}
	input := &CreateCustomerRequest{FirstName: req.GetFirstName(), LastName: req.GetLastName(), Email: req.GetEmail(), Phone: req.GetPhone(), DateOfBirth: req.GetDateOfBirth(), Gender: req.GetGender(), Address: req.GetAddress(), City: req.GetCity(), Country: req.GetCountry(), Tags: req.GetTags(), Notes: req.GetNotes(), Consent: req.GetConsent()}
	if err := handler.Validate(ctx, input); err != nil {
		return nil, err
	}
	result, err := handler.Serve(ctx, input)
	if err != nil {
		return nil, err
	}
	return &customerv1.CreateCustomerResponse{CustomerId: result.CustomerID, Message: result.Message}, nil
}
func (a *CustomerAPI) GetCustomer(ctx context.Context, req *customerv1.GetCustomerRequest) (*customerv1.GetCustomerResponse, error) {
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return nil, err
	}
	handler := &GetCustomerHandler{CRMService: a.service}
	input := &GetCustomerRequest{ID: req.GetId()}
	if err := handler.Validate(ctx, input); err != nil {
		return nil, err
	}
	result, err := handler.Serve(ctx, input)
	if err != nil {
		return nil, err
	}
	return &customerv1.GetCustomerResponse{Customer: result.Customer}, nil
}
func (a *CustomerAPI) UpdateCustomer(ctx context.Context, req *customerv1.UpdateCustomerRequest) (*customerv1.UpdateCustomerResponse, error) {
	if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Customer details are required")
	}
	handler := &UpdateCustomerHandler{CRMService: a.service}
	input := &UpdateCustomerRequest{ID: req.GetId(), FirstName: body.FirstName, LastName: &body.LastName, Email: &body.Email, Phone: &body.Phone, DateOfBirth: &body.DateOfBirth, Gender: &body.Gender, Address: &body.Address, City: &body.City, Country: &body.Country, Notes: &body.Notes, IsActive: body.IsActive}
	if body.Tags != nil {
		input.Tags = &body.Tags
	}
	if body.Consent != nil {
		input.Consent = &body.Consent
	}
	if err := handler.Validate(ctx, input); err != nil {
		return nil, err
	}
	result, err := handler.Serve(ctx, input)
	if err != nil {
		return nil, err
	}
	return &customerv1.UpdateCustomerResponse{Message: result.Message}, nil
}
func (a *CustomerAPI) DeleteCustomer(ctx context.Context, req *customerv1.DeleteCustomerRequest) (*customerv1.DeleteCustomerResponse, error) {
	if err := requireCRMFeaturePermission(ctx, "crm:customers:delete"); err != nil {
		return nil, err
	}
	handler := &DeleteCustomerHandler{CRMService: a.service}
	input := &DeleteCustomerRequest{ID: req.GetId()}
	if err := handler.Validate(ctx, input); err != nil {
		return nil, err
	}
	result, err := handler.Serve(ctx, input)
	if err != nil {
		return nil, err
	}
	return &customerv1.DeleteCustomerResponse{Message: result.Message}, nil
}

type DashboardAPI struct{ service *application.Service }

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{service: service}
}
func (a *DashboardAPI) GetDashboard(ctx context.Context, _ *dashboardv1.GetDashboardRequest) (*dashboardv1.GetDashboardResponse, error) {
	result, err := (&DashboardHandler{CRMService: a.service}).Serve(ctx, &DashboardRequest{})
	if err != nil {
		return nil, err
	}
	return &dashboardv1.GetDashboardResponse{TotalCustomers: int64(result.TotalCustomers), NewThisMonth: int64(result.NewThisMonth), ActiveLoyaltyMembers: int64(result.ActiveLoyaltyMembers), TopCustomers: result.TopCustomers}, nil
}
