package api

import (
	"context"
	"errors"
	"net/http"

	flowviewv1 "github.com/voxire/voxie/gen/go/flow/view/v1"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type ViewAPI struct {
	Service *application.Service
}

func NewViewAPI(service *application.Service) *ViewAPI {
	return &ViewAPI{Service: service}
}

func filtersToStruct(filters map[string]interface{}) *structpb.Struct {
	if filters == nil {
		filters = map[string]interface{}{}
	}
	s, err := structpb.NewStruct(filters)
	if err != nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	return s
}

func filtersFromStruct(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return s.AsMap()
}

func toView(v domain.View) *flowviewv1.View {
	return &flowviewv1.View{
		Id:      v.ID,
		Name:    v.Name,
		Icon:    v.Icon,
		Filters: filtersToStruct(v.Filters),
		GroupBy: v.GroupBy,
		SortBy:  v.SortBy,
		SortDir: v.SortDir,
	}
}

func (a *ViewAPI) ListViews(ctx context.Context, req *flowviewv1.ListViewsRequest) (*flowviewv1.ListViewsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	views, err := a.Service.ListViews(ctx, wsID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}

	out := make([]*flowviewv1.View, 0, len(views))
	for _, v := range views {
		out = append(out, toView(v))
	}
	return &flowviewv1.ListViewsResponse{Views: out}, nil
}

func (a *ViewAPI) CreateView(ctx context.Context, req *flowviewv1.CreateViewRequest) (*flowviewv1.CreateViewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	wsID, err := resolveWorkspaceQuery(ctx, a.Service, authUser, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	filters := filtersFromStruct(req.GetFilters())

	view, err := a.Service.CreateView(ctx, domain.ViewInput{
		WorkspaceID: wsID,
		Name:        req.GetName(),
		Icon:        req.GetIcon(),
		Filters:     filters,
		GroupBy:     req.GetGroupBy(),
		SortBy:      req.GetSortBy(),
		SortDir:     req.GetSortDir(),
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}

	return &flowviewv1.CreateViewResponse{View: toView(view)}, nil
}

func (a *ViewAPI) UpdateView(ctx context.Context, req *flowviewv1.UpdateViewRequest) (*flowviewv1.UpdateViewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.UpdateView(ctx, domain.UpdateViewInput{
		ID:      req.GetId(),
		ActorID: authUser.UserID,
		Name:    req.GetName(),
		Icon:    req.GetIcon(),
		Filters: filtersFromStruct(req.GetFilters()),
		GroupBy: req.GetGroupBy(),
		SortBy:  req.GetSortBy(),
		SortDir: req.GetSortDir(),
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}
	return &flowviewv1.UpdateViewResponse{}, nil
}

func (a *ViewAPI) DeleteView(ctx context.Context, req *flowviewv1.DeleteViewRequest) (*flowviewv1.DeleteViewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteView(ctx, req.GetId(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, err.Error())
	}
	return &flowviewv1.DeleteViewResponse{}, nil
}
