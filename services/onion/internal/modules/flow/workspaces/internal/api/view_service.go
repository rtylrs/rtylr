package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowviewv1 "github.com/voxire/voxie/gen/go/flow/view/v1"
)

type ViewAPI struct {
	Service *application.Service
}

func NewViewAPI(service *application.Service) *ViewAPI {
	return &ViewAPI{Service: service}
}

func filtersToStruct(filters map[string]interface{}) string {
	if filters == nil {
		filters = map[string]interface{}{}
	}
	data, err := json.Marshal(filters)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func filtersFromStruct(s string) map[string]interface{} {
	if s == "" {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}

func toView(v domain.View) *flowviewv1.View {
	return &flowviewv1.View{
		Id:         v.ID,
		Name:       v.Name,
		Icon:       v.Icon,
		Filters:    filtersToStruct(v.Filters),
		GroupBy:    v.GroupBy,
		SortBy:     v.SortBy,
		SortDir:    v.SortDir,
		Visibility: v.Visibility,
	}
}

func (a *ViewAPI) ListViews(ctx context.Context, req *flowviewv1.ListViewsRequest) (*flowviewv1.ListViewsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, businessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	views, total, err := a.Service.ListViews(ctx, businessID, wsID, authUser.UserID, int64(limit), offset)
	if err != nil {
		return nil, workspaceInternalError(ctx, "ListViews failed", err)
	}

	out := make([]*flowviewv1.View, 0, len(views))
	for _, v := range views {
		out = append(out, toView(v))
	}
	return &flowviewv1.ListViewsResponse{
		Views:      out,
		Pagination: buildPagination(page, limit, total, len(views)),
	}, nil
}

func (a *ViewAPI) CreateView(ctx context.Context, req *flowviewv1.CreateViewRequest) (*flowviewv1.CreateViewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, businessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	filters := filtersFromStruct(req.GetFilters())

	view, err := a.Service.CreateView(ctx, domain.ViewInput{
		BusinessID:  businessID,
		WorkspaceID: wsID,
		Name:        req.GetName(),
		Icon:        req.GetIcon(),
		Filters:     filters,
		GroupBy:     req.GetGroupBy(),
		SortBy:      req.GetSortBy(),
		SortDir:     req.GetSortDir(),
		Visibility:  req.GetVisibility(),
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return nil, workspaceInternalError(ctx, "CreateView failed", err)
	}

	return &flowviewv1.CreateViewResponse{View: toView(view)}, nil
}

func (a *ViewAPI) UpdateView(ctx context.Context, req *flowviewv1.UpdateViewRequest) (*flowviewv1.UpdateViewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.UpdateView(ctx, domain.UpdateViewInput{
		ID:         req.GetId(),
		ActorID:    authUser.UserID,
		BusinessID: businessID,
		Name:       req.GetName(),
		Icon:       req.GetIcon(),
		Filters:    filtersFromStruct(req.GetFilters()),
		GroupBy:    req.GetGroupBy(),
		SortBy:     req.GetSortBy(),
		SortDir:    req.GetSortDir(),
		Visibility: req.GetVisibility(),
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, workspaceInternalError(ctx, "UpdateView failed", err)
	}
	return &flowviewv1.UpdateViewResponse{}, nil
}

func (a *ViewAPI) DeleteView(ctx context.Context, req *flowviewv1.DeleteViewRequest) (*flowviewv1.DeleteViewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteView(ctx, req.GetId(), authUser.UserID, businessID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, workspaceInternalError(ctx, "DeleteView failed", err)
	}
	return &flowviewv1.DeleteViewResponse{}, nil
}
