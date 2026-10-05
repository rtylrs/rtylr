package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking/resources/internal/application"
	"github.com/voxire/onion/internal/modules/booking/resources/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingresourcev1 "github.com/voxire/voxie/gen/go/booking/resource/v1"
)

type ResourceAPI struct {
	service *application.Service
}

func NewResourceAPI(service *application.Service) *ResourceAPI {
	return &ResourceAPI{service: service}
}

func (a *ResourceAPI) ListResources(ctx context.Context, _ *bookingresourcev1.ListResourcesRequest) (*bookingresourcev1.ListResourcesResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list resources")
	}
	return &bookingresourcev1.ListResourcesResponse{Resources: newResources(items)}, nil
}

func (a *ResourceAPI) CreateResource(ctx context.Context, req *bookingresourcev1.CreateResourceRequest) (*bookingresourcev1.CreateResourceResponse, error) {
	input, errModel := newResourceInput(req.GetName(), req.GetType(), req.GetCapacity(), req.GetIsActive(), true)
	if errModel != nil {
		return nil, errModel
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, err := a.service.Create(ctx, businessID, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create resource")
	}
	return &bookingresourcev1.CreateResourceResponse{Id: id, Message: "Resource created"}, nil
}

func (a *ResourceAPI) UpdateResource(ctx context.Context, req *bookingresourcev1.UpdateResourceRequest) (*bookingresourcev1.UpdateResourceResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	input, errModel := newResourceInput(req.GetName(), req.GetType(), req.GetCapacity(), req.GetIsActive(), false)
	if errModel != nil {
		return nil, errModel
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.service.Update(ctx, businessID, req.GetId(), input); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update resource")
	}
	return &bookingresourcev1.UpdateResourceResponse{Message: "Resource updated"}, nil
}

func (a *ResourceAPI) DeleteResource(ctx context.Context, req *bookingresourcev1.DeleteResourceRequest) (*bookingresourcev1.DeleteResourceResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete resource")
	}
	return &bookingresourcev1.DeleteResourceResponse{Message: "Resource deleted"}, nil
}

func newResourceInput(name, resourceType string, capacity int32, isActive, requireName bool) (domain.ResourceInput, error) {
	name = strings.TrimSpace(name)
	if requireName && name == "" {
		return domain.ResourceInput{}, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	return domain.ResourceInput{
		Name:     name,
		Type:     resourceType,
		Capacity: int(capacity),
		IsActive: isActive,
	}, nil
}

func newResources(items []domain.Resource) []*bookingresourcev1.Resource {
	out := make([]*bookingresourcev1.Resource, 0, len(items))
	for _, item := range items {
		out = append(out, &bookingresourcev1.Resource{
			Id:         item.ID,
			OwnerId:    item.OwnerID,
			BusinessId: item.BusinessID,
			Name:       item.Name,
			Type:       item.Type,
			Capacity:   int32(item.Capacity), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			IsActive:   item.IsActive,
			CreatedAt:  formatTime(item.CreatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
