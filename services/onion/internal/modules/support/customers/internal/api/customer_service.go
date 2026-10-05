package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/support/customers/internal/application"
	"github.com/voxire/onion/internal/modules/support/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	supportcustomerv1 "github.com/voxire/voxie/gen/go/support/customer/v1"
)

type CustomerAPI struct {
	service *application.Service
}

func NewCustomerAPI(service *application.Service) *CustomerAPI {
	return &CustomerAPI{service: service}
}

func (a *CustomerAPI) ListCustomers(ctx context.Context, req *supportcustomerv1.ListCustomersRequest) (*supportcustomerv1.ListCustomersResponse, error) {
	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	items, total, err := a.service.List(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list support customers")
	}

	return &supportcustomerv1.ListCustomersResponse{
		Customers: newCustomers(items),
		Total:     int32(total), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Page:      int32(page),
		PageSize:  int32(pageSize),
	}, nil
}

func (a *CustomerAPI) GetCustomer(ctx context.Context, req *supportcustomerv1.GetCustomerRequest) (*supportcustomerv1.GetCustomerResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	customer, err := a.service.GetByID(ctx, businessID, req.GetId())
	if errors.Is(err, domain.ErrCustomerNotFound) {
		return nil, onekiterrors.New(http.StatusNotFound, "Support customer not found")
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get support customer")
	}

	return &supportcustomerv1.GetCustomerResponse{Customer: newCustomer(*customer)}, nil
}

func businessIDFromContext(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return device.User.BusinessID, nil
}

func newCustomers(items []domain.Customer) []*supportcustomerv1.Customer {
	out := make([]*supportcustomerv1.Customer, 0, len(items))
	for _, item := range items {
		out = append(out, newCustomer(item))
	}
	return out
}

func newCustomer(item domain.Customer) *supportcustomerv1.Customer {
	lastTicketAt := ""
	if item.LastTicketAt != nil {
		lastTicketAt = formatTime(*item.LastTicketAt)
	}
	return &supportcustomerv1.Customer{
		Id:           item.ID,
		BusinessId:   item.BusinessID,
		Name:         item.Name,
		Email:        item.Email,
		Phone:        item.Phone,
		TotalTickets: int32(item.TotalTickets), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		LastTicketAt: lastTicketAt,
		CreatedAt:    formatTime(item.CreatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
