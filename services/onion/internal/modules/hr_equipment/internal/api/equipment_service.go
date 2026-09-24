package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

type EquipmentAPI struct {
	Service *application.Service
}

func NewEquipmentAPI(service *application.Service) *EquipmentAPI {
	return &EquipmentAPI{Service: service}
}

func (a *EquipmentAPI) ListAssets(ctx context.Context, req *hrequipmentv1.ListAssetsRequest) (*hrequipmentv1.ListAssetsResponse, error) {
	result, err := (&ListAssetsHandler{Service: a.Service}).Serve(ctx, &ListAssetsRequest{
		EmployeeID: req.GetEmployeeId(),
		Condition:  req.GetCondition(),
	})
	if err != nil {
		return nil, err
	}
	return &hrequipmentv1.ListAssetsResponse{
		Assets: result.Assets,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (a *EquipmentAPI) ListEquipmentRequests(ctx context.Context, req *hrequipmentv1.ListEquipmentRequestsRequest) (*hrequipmentv1.ListEquipmentRequestsResponse, error) {
	result, err := (&ListEquipmentRequestsHandler{Service: a.Service}).Serve(ctx, &ListEquipmentRequestsRequest{
		EmployeeID: req.GetEmployeeId(),
		Status:     req.GetStatus(),
		Limit:      50,
	})
	if err != nil {
		return nil, err
	}
	return &hrequipmentv1.ListEquipmentRequestsResponse{
		Requests: result.Requests,
		Total:    result.Total,
		Limit:    result.Limit,
		Offset:   result.Offset,
	}, nil
}

func (a *EquipmentAPI) CreateEquipmentRequest(ctx context.Context, req *hrequipmentv1.CreateEquipmentRequestRequest) (*hrequipmentv1.CreateEquipmentRequestResponse, error) {
	result, err := (&CreateEquipmentRequestHandler{Service: a.Service}).Serve(ctx, &CreateEquipmentRequestRequest{
		EmployeeID:   req.GetEmployeeId(),
		ItemName:     req.GetEquipmentType(),
		Quantity:     1,
		Reason:       req.GetDescription(),
		NeededByDate: req.GetNeededByDate(),
	})
	if err != nil {
		return nil, err
	}
	return &hrequipmentv1.CreateEquipmentRequestResponse{Request: result.Request}, nil
}

func (a *EquipmentAPI) ApproveEquipment(ctx context.Context, req *hrequipmentv1.ApproveEquipmentRequest) (*hrequipmentv1.ApproveEquipmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	// An absent body means the client never sent a decision. Silently reading
	// the zero value as "rejected" used to persist a rejection for a valid
	// pending request while responding success:true.
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "request body with the approval decision is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	approved := body.GetApproved()
	if err := a.Service.ApproveEquipmentRequest(ctx, businessID, req.GetEquipmentRequestId(), device.User.UserID, approved, ""); err != nil {
		logger.Ctx(ctx).Error("failed to process equipment request", logger.Any("owner_id", device.User.UserID), logger.Any("request_id", req.GetEquipmentRequestId()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to process equipment request")
	}
	return &hrequipmentv1.ApproveEquipmentResponse{Success: true}, nil
}
