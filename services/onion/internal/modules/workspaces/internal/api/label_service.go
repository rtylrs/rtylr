package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowlabelv1 "github.com/voxire/voxie/gen/go/flow/label/v1"
)

type LabelAPI struct {
	Service *application.Service
}

func NewLabelAPI(service *application.Service) *LabelAPI {
	return &LabelAPI{Service: service}
}

func toLabel(l domain.Label) *flowlabelv1.Label {
	return &flowlabelv1.Label{
		Id:          l.ID,
		WorkspaceId: l.WorkspaceID,
		Name:        l.Name,
		Color:       l.Color,
		CreatedAt:   l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (a *LabelAPI) resolveWorkspaceID(ctx context.Context, userID, businessID, workspaceID int64) (int64, error) {
	raw := strconv.FormatInt(workspaceID, 10)
	if workspaceID == 0 {
		raw = "personal"
	}
	resolved, err := a.Service.ResolveWorkspaceID(ctx, userID, businessID, raw)
	if err != nil || resolved == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}
	return resolved, nil
}

func (a *LabelAPI) ListLabels(ctx context.Context, req *flowlabelv1.ListLabelsRequest) (*flowlabelv1.ListLabelsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	workspaceID, err := a.resolveWorkspaceID(ctx, authUser.UserID, authUser.BusinessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	labels, total, err := a.Service.ListLabels(ctx, workspaceID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list labels")
	}

	out := make([]*flowlabelv1.Label, 0, len(labels))
	for _, l := range labels {
		out = append(out, toLabel(l))
	}
	return &flowlabelv1.ListLabelsResponse{
		Labels:     out,
		Pagination: buildPagination(page, limit, total, len(labels)),
	}, nil
}

func (a *LabelAPI) CreateLabel(ctx context.Context, req *flowlabelv1.CreateLabelRequest) (*flowlabelv1.CreateLabelResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	workspaceID, err := a.resolveWorkspaceID(ctx, authUser.UserID, authUser.BusinessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}

	label, err := a.Service.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: workspaceID,
		Name:        req.GetName(),
		Color:       req.GetColor(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create label")
	}

	return &flowlabelv1.CreateLabelResponse{Label: toLabel(label)}, nil
}

func (a *LabelAPI) UpdateLabel(ctx context.Context, req *flowlabelv1.UpdateLabelRequest) (*flowlabelv1.UpdateLabelResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.UpdateLabel(ctx, req.GetId(), authUser.UserID, businessID, domain.LabelInput{
		Name:  req.GetName(),
		Color: req.GetColor(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update label")
	}

	return &flowlabelv1.UpdateLabelResponse{Message: "Label updated"}, nil
}

func (a *LabelAPI) DeleteLabel(ctx context.Context, req *flowlabelv1.DeleteLabelRequest) (*flowlabelv1.DeleteLabelResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteLabel(ctx, req.GetId(), authUser.UserID, businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete label")
	}

	return &flowlabelv1.DeleteLabelResponse{Message: "Label deleted"}, nil
}

func (a *LabelAPI) SetTaskLabels(ctx context.Context, req *flowlabelv1.SetTaskLabelsRequest) (*flowlabelv1.SetTaskLabelsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	// Setting a task's labels is a task mutation, so it takes the same tier as
	// the handlers in task_service.go. The repository guard below only checks
	// workspace membership, which admits a viewer.
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	labelIDs := req.GetLabelIds()
	if labelIDs == nil {
		labelIDs = []int64{}
	}

	if err := a.Service.SetTaskLabels(ctx, req.GetTaskId(), authUser.UserID, labelIDs); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to set task labels")
	}

	return &flowlabelv1.SetTaskLabelsResponse{Message: "Task labels updated"}, nil
}
