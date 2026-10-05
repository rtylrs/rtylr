package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/application"
	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hroneononev1 "github.com/voxire/voxie/gen/go/hr/one_on_one/v1"
)

var scheduledAtLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

type OneOnOneAPI struct {
	service *application.Service
}

func NewOneOnOneAPI(service *application.Service) *OneOnOneAPI {
	return &OneOnOneAPI{service: service}
}

func (a *OneOnOneAPI) ListOneOnOnes(ctx context.Context, _ *hroneononev1.ListOneOnOnesRequest) (*hroneononev1.ListOneOnOnesResponse, error) {
	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list one-on-ones")
	}

	return &hroneononev1.ListOneOnOnesResponse{OneOnOnes: newOneOnOnes(items)}, nil
}

func (a *OneOnOneAPI) CreateOneOnOne(ctx context.Context, req *hroneononev1.CreateOneOnOneRequest) (*hroneononev1.CreateOneOnOneResponse, error) {
	if req.GetManagerId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "managerId is required")
	}
	if req.GetStaffId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "staffId is required")
	}
	if req.GetScheduledAt() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "scheduledAt is required")
	}
	if !isValidScheduledAt(req.GetScheduledAt()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "scheduledAt is not a valid date")
	}

	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	id, err := a.service.Create(ctx, businessID, domain.OneOnOneInput{
		ManagerID:   req.GetManagerId(),
		StaffID:     req.GetStaffId(),
		ScheduledAt: req.GetScheduledAt(),
		Notes:       req.GetNotes(),
		ActionItems: req.GetActionItems(),
		Status:      req.GetStatus(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create one-on-one")
	}

	return &hroneononev1.CreateOneOnOneResponse{Id: id, Message: "One-on-one created successfully"}, nil
}

func resolveBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func newOneOnOnes(items []domain.OneOnOne) []*hroneononev1.OneOnOne {
	out := make([]*hroneononev1.OneOnOne, 0, len(items))
	for _, item := range items {
		out = append(out, &hroneononev1.OneOnOne{
			Id:          item.ID,
			OwnerId:     item.OwnerID,
			ManagerId:   item.ManagerID,
			StaffId:     item.StaffID,
			ScheduledAt: formatTime(item.ScheduledAt),
			Notes:       item.Notes,
			ActionItems: item.ActionItems,
			Status:      item.Status,
			CreatedAt:   formatTime(item.CreatedAt),
		})
	}
	return out
}

func isValidScheduledAt(value string) bool {
	for _, layout := range scheduledAtLayouts {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}
	return false
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
