package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

type OnboardingService struct {
	service *application.Service
}

func NewOnboardingService(service *application.Service) *OnboardingService {
	return &OnboardingService{service: service}
}

func (s *OnboardingService) ListChecklistTemplates(ctx context.Context, req *hronboardingv1.ListChecklistTemplatesRequest) (*hronboardingv1.ListChecklistTemplatesResponse, error) {
	if req == nil {
		req = &hronboardingv1.ListChecklistTemplatesRequest{}
	}
	res, err := (&ListTemplatesHandler{Service: s.service}).Serve(ctx, &ListTemplatesRequest{
		TriggerType: req.TriggerType,
	})
	if err != nil {
		return nil, err
	}
	return &hronboardingv1.ListChecklistTemplatesResponse{
		Templates: res.Templates,
		Total:     res.Total,
		Limit:     res.Limit,
		Offset:    res.Offset,
	}, nil
}

func (s *OnboardingService) CreateChecklistTemplate(ctx context.Context, req *hronboardingv1.CreateChecklistTemplateRequest) (*hronboardingv1.CreateChecklistTemplateResponse, error) {
	in := &CreateTemplateRequest{Name: req.GetName(), TriggerType: req.GetTriggerType(), RoleID: req.GetRoleId()}
	handler := &CreateTemplateHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hronboardingv1.CreateChecklistTemplateResponse{Template: res.Template}, nil
}

func (s *OnboardingService) UpdateChecklistTemplate(ctx context.Context, req *hronboardingv1.UpdateChecklistTemplateRequest) (*hronboardingv1.UpdateChecklistTemplateResponse, error) {
	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.Name) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if err := requireOnboardingManagePermission(ctx); err != nil {
		return nil, err
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	err := s.service.UpdateTemplate(ctx, businessID, req.GetRoleId(), domain.UpdateTemplateParams{
		Name:        body.Name,
		TriggerType: body.TriggerType,
	})
	if err != nil {
		if errors.Is(err, domain.ErrChecklistAccessDenied) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Template not found")
		}
		if errors.Is(err, domain.ErrEmptyTemplateName) || errors.Is(err, domain.ErrInvalidTriggerType) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		logger.Ctx(ctx).Error("failed to update onboarding template", logger.Any("owner_id", device.User.UserID), logger.Any("template_id", req.GetRoleId()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update template")
	}
	return &hronboardingv1.UpdateChecklistTemplateResponse{Success: true}, nil
}

func (s *OnboardingService) AddItemToChecklistTemplate(ctx context.Context, req *hronboardingv1.AddItemToChecklistTemplateRequest) (*hronboardingv1.AddItemToChecklistTemplateResponse, error) {
	if err := requireOnboardingManagePermission(ctx); err != nil {
		return nil, err
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	item, err := s.service.AddItemToTemplate(ctx, businessID, req.GetRoleId(), domain.CreateTemplateItemParams{
		Title:              req.GetTitle(),
		Description:        req.GetDescription(),
		AssigneeRole:       req.GetAssigneeRole(),
		DueOffsetDays:      req.GetDueOffsetDays(),
		RequiresAttachment: req.GetRequiresAttachment(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrChecklistAccessDenied) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Template not found")
		}
		logger.Ctx(ctx).Error("failed to add item to onboarding template", logger.Any("owner_id", device.User.UserID), logger.Any("template_id", req.GetRoleId()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add item to template")
	}
	return &hronboardingv1.AddItemToChecklistTemplateResponse{Item: checklistItemToProto(*item)}, nil
}

func (s *OnboardingService) ListEmployeeChecklists(ctx context.Context, req *hronboardingv1.ListEmployeeChecklistsRequest) (*hronboardingv1.ListEmployeeChecklistsResponse, error) {
	res, err := (&ListEmployeeChecklistsHandler{Service: s.service}).Serve(ctx, &ListEmployeeChecklistsRequest{
		EmployeeID: req.GetEmployeeId(),
		Limit:      req.GetLimit(),
		Offset:     req.GetOffset(),
	})
	if err != nil {
		return nil, err
	}
	return &hronboardingv1.ListEmployeeChecklistsResponse{
		Checklists: res.Checklists,
		Total:      res.Total,
		Limit:      res.Limit,
		Offset:     res.Offset,
	}, nil
}

func (s *OnboardingService) GetEmployeeChecklist(ctx context.Context, req *hronboardingv1.GetEmployeeChecklistRequest) (*hronboardingv1.GetEmployeeChecklistResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	checklist, err := s.service.GetLatestEmployeeChecklist(ctx, businessID, req.GetEmployeeId())
	if err != nil {
		if errors.Is(err, domain.ErrChecklistNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Checklist not found")
		}
		logger.Ctx(ctx).Error("failed to get employee checklist", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", req.GetEmployeeId()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get checklist")
	}
	return &hronboardingv1.GetEmployeeChecklistResponse{Checklist: employeeChecklistToProto(*checklist)}, nil
}

func (s *OnboardingService) TriggerChecklist(ctx context.Context, req *hronboardingv1.TriggerChecklistRequest) (*hronboardingv1.TriggerChecklistResponse, error) {
	in := &TriggerChecklistRequest{EmployeeID: req.GetEmployeeId(), TemplateID: req.GetTemplateId(), TriggerDate: req.GetTriggerDate()}
	handler := &TriggerChecklistHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hronboardingv1.TriggerChecklistResponse{Checklist: res.Checklist}, nil
}

func (s *OnboardingService) CompleteChecklistItem(_ context.Context, _ *hronboardingv1.CompleteChecklistItemRequest) (*hronboardingv1.CompleteChecklistItemResponse, error) {
	return nil, onekiterrors.New(http.StatusBadRequest, "checklist item id is required")
}
