package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
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

type CustomerPayload struct {
	ID            int64    `json:"id"`
	OwnerID       int64    `json:"owner_id"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	DateOfBirth   *string  `json:"date_of_birth"`
	Gender        string   `json:"gender"`
	Address       string   `json:"address"`
	City          string   `json:"city"`
	Country       string   `json:"country"`
	LoyaltyPoints int      `json:"loyalty_points"`
	TotalSpent    float64  `json:"total_spent"`
	VisitCount    int      `json:"visit_count"`
	LastVisitAt   *string  `json:"last_visit_at"`
	LoyaltyTier   string   `json:"loyalty_tier"`
	Tags          []string `json:"tags"`
	Notes         string   `json:"notes"`
	IsActive      bool     `json:"is_active"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type ListCustomersResponse struct {
	Customers []CustomerPayload `json:"customers"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"page_size"`
}

func newCustomerPayload(c domain.Customer) CustomerPayload {
	p := CustomerPayload{
		ID:            c.ID,
		OwnerID:       c.OwnerID,
		FirstName:     c.FirstName,
		LastName:      c.LastName,
		Email:         c.Email,
		Phone:         c.Phone,
		Gender:        c.Gender,
		Address:       c.Address,
		City:          c.City,
		Country:       c.Country,
		LoyaltyPoints: c.LoyaltyPoints,
		TotalSpent:    c.TotalSpent,
		VisitCount:    c.VisitCount,
		LoyaltyTier:   c.LoyaltyTier,
		Tags:          c.Tags,
		Notes:         c.Notes,
		IsActive:      c.IsActive,
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

func (h *ListCustomersHandler) Serve(ctx context.Context, req *ListCustomersRequest) (ListCustomersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.PageSize < 1 {
		req.PageSize = 20
	}

	result, err := h.CRMService.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:  device.User.UserID,
		Search:   req.Search,
		Tier:     req.Tier,
		Tag:      req.Tag,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return ListCustomersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list customers",
		}
	}

	payloads := make([]CustomerPayload, len(result.Customers))
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
