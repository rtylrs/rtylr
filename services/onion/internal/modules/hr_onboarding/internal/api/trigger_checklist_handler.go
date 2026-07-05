package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func TriggerChecklistRoutes() []string { return []string{"/hr/onboarding/checklists/trigger"} }

type TriggerChecklistHandler struct{ Service *application.Service }

type TriggerChecklistRequest struct {
	EmployeeID  int64  `json:"employee_id"`
	TemplateID  int64  `json:"template_id"`
	TriggerDate string `json:"trigger_date"`
}

type TriggerChecklistResponse struct {
	Checklist *hronboardingv1.EmployeeChecklistPayload `json:"checklist"`
}

func (h *TriggerChecklistHandler) Validate(_ context.Context, req *TriggerChecklistRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.TemplateID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "template_id is required")
	}
	if req.TriggerDate == "" {
		return onekiterrors.New(http.StatusBadRequest, "trigger_date is required")
	}
	if _, err := time.Parse("2006-01-02", req.TriggerDate); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "trigger_date must be in YYYY-MM-DD format")
	}
	return nil
}

func (h *TriggerChecklistHandler) Serve(ctx context.Context, req *TriggerChecklistRequest) (TriggerChecklistResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return TriggerChecklistResponse{}, onekiterrors.PermissionDenied()
	}

	triggerDate, err := time.Parse("2006-01-02", req.TriggerDate)
	if err != nil {
		return TriggerChecklistResponse{}, onekiterrors.New(http.StatusBadRequest, "invalid trigger_date format, expected YYYY-MM-DD")
	}

	checklist, err := h.Service.TriggerChecklist(ctx, device.User.UserID, req.EmployeeID, req.TemplateID, triggerDate)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return TriggerChecklistResponse{}, onekiterrors.New(http.StatusNotFound, "Template not found")
		}
		logger.Ctx(ctx).Error("failed to trigger onboarding checklist", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return TriggerChecklistResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to trigger checklist")
	}

	return TriggerChecklistResponse{Checklist: employeeChecklistToProto(*checklist)}, nil
}
