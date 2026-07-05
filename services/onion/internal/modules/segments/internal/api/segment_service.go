package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	crmsegmentv1 "github.com/voxire/voxie/gen/go/crm/segment/v1"

	"github.com/voxire/onion/internal/modules/segments/internal/application"
	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type SegmentAPI struct {
	service *application.Service
}

func NewSegmentAPI(service *application.Service) *SegmentAPI {
	return &SegmentAPI{service: service}
}

func (a *SegmentAPI) ListSegments(ctx context.Context, _ *crmsegmentv1.ListSegmentsRequest) (*crmsegmentv1.ListSegmentsResponse, error) {
	businessID, _, errModel := resolveRequestIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	segments, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list segments")
	}

	return &crmsegmentv1.ListSegmentsResponse{Segments: newSegments(segments)}, nil
}

func (a *SegmentAPI) CreateSegment(ctx context.Context, req *crmsegmentv1.CreateSegmentRequest) (*crmsegmentv1.CreateSegmentResponse, error) {
	if req.GetName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	businessID, ownerID, errModel := resolveRequestIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	id, err := a.service.Create(ctx, domain.SegmentInput{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Filters:     filtersBytes(req.GetFilters()),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create segment")
	}

	return &crmsegmentv1.CreateSegmentResponse{SegmentId: id, Message: "Segment created successfully"}, nil
}

func (a *SegmentAPI) UpdateSegment(ctx context.Context, req *crmsegmentv1.UpdateSegmentRequest) (*crmsegmentv1.UpdateSegmentResponse, error) {
	body := req.GetBody()
	if body == nil || body.GetName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	businessID, ownerID, errModel := resolveRequestIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	err := a.service.Update(ctx, domain.SegmentUpdateInput{
		ID:          req.GetId(),
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Name:        body.GetName(),
		Description: body.GetDescription(),
		Filters:     filtersBytes(body.GetFilters()),
	})
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update segment")
	}

	return &crmsegmentv1.UpdateSegmentResponse{Message: "Segment updated successfully"}, nil
}

func (a *SegmentAPI) DeleteSegment(ctx context.Context, req *crmsegmentv1.DeleteSegmentRequest) (*crmsegmentv1.DeleteSegmentResponse, error) {
	businessID, _, errModel := resolveRequestIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	err := a.service.Delete(ctx, businessID, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete segment")
	}

	return &crmsegmentv1.DeleteSegmentResponse{Message: "Segment deleted successfully"}, nil
}

func (a *SegmentAPI) GetSegmentDetail(context.Context, *crmsegmentv1.GetSegmentDetailRequest) (*crmsegmentv1.GetSegmentDetailResponse, error) {
	return nil, notImplemented("Get segment detail")
}

func (a *SegmentAPI) GetSegmentCustomers(context.Context, *crmsegmentv1.GetSegmentCustomersRequest) (*crmsegmentv1.GetSegmentCustomersResponse, error) {
	return nil, notImplemented("Get segment customers")
}

func resolveRequestIdentity(ctx context.Context) (businessID int64, ownerID int64, errModel error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	return businessID, device.User.UserID, nil
}

func newSegments(items []domain.Segment) []*crmsegmentv1.Segment {
	out := make([]*crmsegmentv1.Segment, 0, len(items))
	for _, item := range items {
		out = append(out, &crmsegmentv1.Segment{
			Id:            item.ID,
			OwnerId:       item.OwnerID,
			Name:          item.Name,
			Description:   item.Description,
			Filters:       filtersString(item.Filters),
			CustomerCount: int64(item.CustomerCount),
			CreatedAt:     formatTime(item.CreatedAt),
			UpdatedAt:     formatTime(item.UpdatedAt),
		})
	}
	return out
}

func filtersString(filters []byte) string {
	if len(filters) == 0 {
		return "{}"
	}
	return string(filters)
}

func filtersBytes(filters string) []byte {
	if filters == "" {
		return []byte("{}")
	}
	return []byte(filters)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func notImplemented(method string) error {
	return onekiterrors.New(http.StatusNotImplemented, method+" is not implemented on OneKit yet")
}
