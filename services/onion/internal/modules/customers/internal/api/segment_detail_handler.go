package api

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
	crmsegmentv1 "github.com/voxire/voxie/gen/go/crm/segment/v1"
)

func SegmentDetailRoutes() []string {
	return []string{"/crm/segments/detail"}
}

func SegmentCustomersRoutes() []string {
	return []string{"/crm/segments/customers"}
}

// -- Segment Detail ----------------------------------------------------------

type GetSegmentDetailHandler struct {
	CRMService *application.Service
}

type GetSegmentDetailRequest struct {
	ID int64 `json:"id"`
}

type GetSegmentDetailResponse struct {
	Segment   *crmsegmentv1.Segment     `json:"segment"`
	Customers []*crmcustomerv1.Customer `json:"customers"`
}

func (h *GetSegmentDetailHandler) Validate(_ context.Context, req *GetSegmentDetailRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Segment ID is required")
	}

	return nil
}

func (h *GetSegmentDetailHandler) Serve(ctx context.Context, req *GetSegmentDetailRequest) (GetSegmentDetailResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSegmentDetailResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetSegmentDetailResponse{}, onekiterrors.PermissionDenied()
	}

	detail, err := h.CRMService.GetSegmentDetail(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return GetSegmentDetailResponse{}, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}

		return GetSegmentDetailResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get segment detail")
	}

	custPayloads := make([]*crmcustomerv1.Customer, len(detail.Customers))
	for i, c := range detail.Customers {
		custPayloads[i] = newCustomerPayload(c)
	}

	return GetSegmentDetailResponse{
		Segment:   newSegmentPayload(detail.Segment),
		Customers: custPayloads,
	}, nil
}

// -- Segment Customers -------------------------------------------------------

type GetSegmentCustomersHandler struct {
	CRMService *application.Service
}

type GetSegmentCustomersRequest struct {
	SegmentID int64 `json:"segment_id"`
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
}

type GetSegmentCustomersResponse struct {
	Customers []*crmcustomerv1.Customer `json:"customers"`
	Total     int                       `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
}

const (
	segmentCustomersDefaultPageSize = 20
	segmentCustomersMaxPageSize     = 100
)

func normalizeSegmentCustomersPage(page, pageSize int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = segmentCustomersDefaultPageSize
	}
	if pageSize > segmentCustomersMaxPageSize {
		pageSize = segmentCustomersMaxPageSize
	}
	if int64(page-1) > math.MaxInt32/int64(pageSize) {
		return 0, 0, onekiterrors.New(http.StatusBadRequest, "page offset is too large")
	}
	return page, pageSize, nil
}

func (h *GetSegmentCustomersHandler) Validate(_ context.Context, req *GetSegmentCustomersRequest) error {
	if req.SegmentID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Segment ID is required")
	}

	return nil
}

func (h *GetSegmentCustomersHandler) Serve(ctx context.Context, req *GetSegmentCustomersRequest) (GetSegmentCustomersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSegmentCustomersResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetSegmentCustomersResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		return GetSegmentCustomersResponse{}, onekiterrors.New(http.StatusBadRequest, "Segment customer request is required")
	}

	page, pageSize, err := normalizeSegmentCustomersPage(req.Page, req.PageSize)
	if err != nil {
		return GetSegmentCustomersResponse{}, err
	}
	req.Page = page
	req.PageSize = pageSize

	customers, total, err := h.CRMService.GetSegmentCustomers(ctx, businessID, req.SegmentID, req.Page, req.PageSize)
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return GetSegmentCustomersResponse{}, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}

		return GetSegmentCustomersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get segment customers")
	}

	payloads := make([]*crmcustomerv1.Customer, len(customers))
	for i, c := range customers {
		payloads[i] = newCustomerPayload(c)
	}

	return GetSegmentCustomersResponse{
		Customers: payloads,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}
