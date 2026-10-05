package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubteamaccessv1 "github.com/voxire/voxie/gen/go/hub/team/access/v1"
)

type AccessAPI struct {
	Service *application.Service
}

type ListHierarchyRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type ListHierarchyResponse struct {
	Members []staffPayload `json:"members"`
	HasMore bool           `json:"has_more"`
}

func NewAccessAPI(service *application.Service) *AccessAPI {
	return &AccessAPI{Service: service}
}

func (a *AccessAPI) RequestAccess(ctx context.Context, req *hubteamaccessv1.RequestAccessRequest) (*hubteamaccessv1.RequestAccessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	app := req.GetApp()
	if err := a.Service.RequestAccess(ctx, device.User.BusinessID, device.User.UserID, app); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to request access")
	}

	requesterEmail, requesterName, _ := a.Service.GetUserInfo(ctx, device.User.UserID)
	if requesterName == "" {
		requesterName = requesterEmail
	}

	a.Service.NotifyAdminsOfAccessRequest(ctx, device.User.BusinessID, requesterName, app)
	if emailErr := a.Service.EnqueueAccessRequestEmail(ctx, device.User.BusinessID, requesterName, requesterEmail, app); emailErr != nil {
		logger.Ctx(ctx).Warn("failed to enqueue access request email", logger.Error(emailErr))
	}

	return &hubteamaccessv1.RequestAccessResponse{Message: "Access requested"}, nil
}

func (a *AccessAPI) ListAccessRequests(ctx context.Context, _ *hubteamaccessv1.ListAccessRequestsRequest) (*hubteamaccessv1.ListAccessRequestsResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:view")
	if permissionErr != nil {
		return nil, permissionErr
	}

	reqs, err := a.Service.ListAccessRequests(ctx, device.UserID, device.BusinessID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list access requests")
	}

	payloads := make([]*hubteamaccessv1.AccessRequest, 0, len(reqs))
	for _, r := range reqs {
		payloads = append(payloads, &hubteamaccessv1.AccessRequest{
			Id:             r.ID,
			UserId:         r.UserID,
			App:            r.App,
			RequesterName:  r.RequesterName,
			RequesterEmail: r.RequesterEmail,
			CreatedAt:      r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &hubteamaccessv1.ListAccessRequestsResponse{Requests: payloads}, nil
}

func (a *AccessAPI) ApproveAccessRequest(ctx context.Context, req *hubteamaccessv1.ApproveAccessRequestRequest) (*hubteamaccessv1.ApproveAccessRequestResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:edit")
	if permissionErr != nil {
		return nil, permissionErr
	}

	requestID := req.GetId()
	if requestID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid request id")
	}
	userID := req.GetUserId()
	app := req.GetApp()
	if userID <= 0 || app == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "user_id and app are required")
	}

	if err := a.Service.ApproveAccessRequest(ctx, device.UserID, device.BusinessID, requestID, userID, app); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to approve access request")
	}

	return &hubteamaccessv1.ApproveAccessRequestResponse{Message: "Access granted"}, nil
}

func (a *AccessAPI) ListHierarchy(ctx context.Context, req *ListHierarchyRequest) (ListHierarchyResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:hierarchy:manage")
	if permissionErr != nil {
		return ListHierarchyResponse{}, permissionErr
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.BusinessID, device.UserID)
	if businessErr != nil {
		return ListHierarchyResponse{}, onekiterrors.PermissionDenied()
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	members, hasMore, err := a.Service.ListStaffUsersByBusiness(ctx, businessID, page, pageSize)
	if err != nil {
		return ListHierarchyResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list hierarchy")
	}
	return ListHierarchyResponse{Members: staffPayloads(members), HasMore: hasMore}, nil
}
