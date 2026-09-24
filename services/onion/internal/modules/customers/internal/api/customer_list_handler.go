package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
)

const (
	defaultCustomerPageSize = 20
	maxCustomerPageSize     = 100
)

func CustomerListRoutes() []string {
	return []string{"/crm/customers"}
}

type ListCustomersHandler struct {
	CRMService *application.Service
}

type ListCustomersRequest struct {
	Search   string `json:"search"`
	Tier     string `json:"tier"`
	Tag      string `json:"tag"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ListCustomersResponse struct {
	Customers []*crmcustomerv1.Customer `json:"customers"`
	Total     int                       `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
}

func newCustomerPayload(c domain.Customer) *crmcustomerv1.Customer {
	p := &crmcustomerv1.Customer{
		Id:            c.ID,
		OwnerId:       c.OwnerID,
		FirstName:     c.FirstName,
		LastName:      c.LastName,
		Email:         c.Email,
		Phone:         c.Phone,
		Gender:        c.Gender,
		Address:       c.Address,
		City:          c.City,
		Country:       c.Country,
		LoyaltyPoints: int64(c.LoyaltyPoints),
		TotalSpent:    c.TotalSpent,
		VisitCount:    int64(c.VisitCount),
		LoyaltyTier:   c.LoyaltyTier,
		Tags:          c.Tags,
		Notes:         c.Notes,
		IsActive:      c.IsActive,
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if len(c.Consent) > 0 {
		_ = json.Unmarshal(c.Consent, &p.Consent)
	}

	if c.DateOfBirth != nil {
		s := c.DateOfBirth.Format("2006-01-02")
		p.DateOfBirth = &s
	}

	if c.LastVisitAt != nil {
		s := c.LastVisitAt.Format("2006-01-02T15:04:05Z")
		p.LastVisitAt = &s
	}

	return p
}

func (h *ListCustomersHandler) Serve(ctx context.Context, req *ListCustomersRequest) (ListCustomersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCustomersResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListCustomersResponse{}, onekiterrors.PermissionDenied()
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.PageSize < 1 || req.PageSize > maxCustomerPageSize {
		req.PageSize = defaultCustomerPageSize
	}

	// The repository binds LIMIT/OFFSET as int32 for MySQL. Keep the computed
	// offset within that range so a very large page number cannot wrap to a
	// different page or produce a negative OFFSET.
	maxPage := math.MaxInt32/req.PageSize + 1
	if req.Page > maxPage {
		req.Page = maxPage
	}

	result, err := h.CRMService.ListCustomers(ctx, domain.CustomerListFilter{
		BusinessID: businessID,
		Search:     req.Search,
		Tier:       req.Tier,
		Tag:        req.Tag,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return ListCustomersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list customers")
	}

	payloads := make([]*crmcustomerv1.Customer, len(result.Customers))
	for i, c := range result.Customers {
		payloads[i] = newCustomerPayload(c)
	}

	return ListCustomersResponse{
		Customers: payloads,
		Total:     result.Total,
		Page:      result.Page,
		PageSize:  result.PageSize,
	}, nil
}
