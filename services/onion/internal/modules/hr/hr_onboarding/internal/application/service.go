package application

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrOnboardingScope = "hr_onboarding"
	hrOnboardingTTL   = 5 * time.Minute
)

type repository interface {
	CreateTemplate(ctx context.Context, businessID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error)
	GetTemplate(ctx context.Context, businessID, id int64) (*domain.ChecklistTemplateRecord, error)
	ListTemplatesFiltered(ctx context.Context, businessID int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error)
	UpdateTemplate(ctx context.Context, businessID, id int64, params domain.UpdateTemplateParams) error
	CreateTemplateItem(ctx context.Context, businessID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error)
	ListTemplateItems(ctx context.Context, businessID, templateID int64) ([]domain.ChecklistItemRecord, error)
	TriggerChecklist(ctx context.Context, businessID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error)
	GetEmployeeChecklist(ctx context.Context, businessID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error)
	GetLatestEmployeeChecklist(ctx context.Context, businessID, employeeID int64) (*domain.EmployeeChecklistRecord, error)
	ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.EmployeeChecklistRecord, int64, error)
	UpdateEmployeeChecklistStatus(ctx context.Context, businessID, checklistID int64, status string) error
	CreateEmployeeChecklistItem(ctx context.Context, businessID, employeeChecklistID, checklistItemID, assignedToEmployeeID int64, dueDate string) error
	GetChecklistItem(ctx context.Context, businessID, itemID int64) (*domain.EmployeeChecklistItemRecord, error)
	CompleteChecklistItemWithAttachment(ctx context.Context, businessID, itemID, completedBy int64, attachmentURL *string) error
	GetChecklistCompletionStats(ctx context.Context, businessID, checklistID int64) (total, completed int, err error)
	ListEmployeeChecklistItems(ctx context.Context, businessID, employeeChecklistID int64) ([]domain.EmployeeChecklistItemRecord, error)
	EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error)
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func invalidateHROnboardingScope(ctx context.Context, businessID int64) {
	pattern := cacheKeys.ScopedPrefix(hrOnboardingScope, businessID)
	cacher.InvalidateScope(ctx, pattern)

	// InvalidateScope clears the writer's local cache and Redis, but other API
	// replicas keep their in-memory entries. The API subscribes to this channel
	// on startup, so publish the same business-scoped pattern for every replica.
	cacher.NewPubSub[cacheInvalidate.PatternInvalidationConsumer]().Publish(
		cacheInvalidate.PatternInvalidationChannel,
		cacheInvalidate.PatternInvalidationConsumer{Patterns: []string{pattern}},
	)
}

func hasOnboardingManagePermission(ctx context.Context, businessID ...int64) bool {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil {
		return false
	}
	if len(businessID) > 0 && businessID[0] > 0 {
		if user.BusinessID > 0 && user.BusinessID != businessID[0] {
			return false
		}
		if user.EffectiveAccess != nil && user.EffectiveAccess.BusinessID > 0 && user.EffectiveAccess.BusinessID != businessID[0] {
			return false
		}
	}
	return (authPkg.HasAppPermission(user, authorization.AppHR, "hr:employees:manage") || authPkg.IsBusinessAdmin(user)) &&
		authPkg.HasAnyAppScope(user, authorization.AppHR)
}

type onboardingListResult[T any] struct {
	Items []T
	Total int64
}

func (s *Service) ListTemplates(ctx context.Context, businessID int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error) {
	if businessID <= 0 {
		return nil, 0, domain.ErrChecklistAccessDenied
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	triggerKey := "*"
	if params.TriggerType != nil {
		triggerKey = url.QueryEscape(*params.TriggerType)
	}
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("tmpl|trigger=%s|l=%d|o=%d", triggerKey, params.Limit, params.Offset))
	res, err := cacher.Of[onboardingListResult[domain.ChecklistTemplateRecord]]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (onboardingListResult[domain.ChecklistTemplateRecord], error) {
			items, total, err := s.repo.ListTemplatesFiltered(ctx, businessID, params)
			return onboardingListResult[domain.ChecklistTemplateRecord]{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateTemplate(ctx context.Context, businessID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	if businessID <= 0 || !hasOnboardingManagePermission(ctx, businessID) {
		return nil, domain.ErrChecklistAccessDenied
	}
	params.Name = strings.TrimSpace(params.Name)
	if params.TriggerType == "" {
		params.TriggerType = domain.TriggerTypeOnboarding
	}
	if err := domain.ValidateCreateTemplateParams(params); err != nil {
		return nil, err
	}

	rec, err := s.repo.CreateTemplate(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(context.WithoutCancel(ctx), businessID)
	return rec, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, businessID, id int64, params domain.UpdateTemplateParams) error {
	if businessID <= 0 || !hasOnboardingManagePermission(ctx, businessID) {
		return domain.ErrChecklistAccessDenied
	}
	existing, err := s.repo.GetTemplate(ctx, businessID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrTemplateNotFound
	}

	params.Name = strings.TrimSpace(params.Name)
	if params.Name == "" {
		return domain.ErrEmptyTemplateName
	}
	if params.TriggerType != nil {
		triggerType := strings.TrimSpace(*params.TriggerType)
		if triggerType == "" {
			return domain.ErrInvalidTriggerType
		}
		params.TriggerType = &triggerType
	}
	if params.TriggerType == nil {
		// The SQL uses COALESCE for nullable patch fields, but carrying the
		// current value here also keeps the repository contract explicit.
		current := existing.TriggerType
		params.TriggerType = &current
	}
	if err := domain.ValidateCreateTemplateParams(domain.CreateTemplateParams{
		Name: params.Name, TriggerType: *params.TriggerType,
	}); err != nil {
		return err
	}

	if err := s.repo.UpdateTemplate(ctx, businessID, id, params); err != nil {
		return err
	}
	invalidateHROnboardingScope(context.WithoutCancel(ctx), businessID)
	return nil
}

func (s *Service) AddItemToTemplate(ctx context.Context, businessID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	if businessID <= 0 || !hasOnboardingManagePermission(ctx, businessID) {
		return nil, domain.ErrChecklistAccessDenied
	}
	if _, err := s.repo.GetTemplate(ctx, businessID, templateID); err != nil {
		return nil, err
	}
	if params.AssigneeRole == "" {
		params.AssigneeRole = "employee"
	}
	rec, err := s.repo.CreateTemplateItem(ctx, businessID, templateID, params)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(context.WithoutCancel(ctx), businessID)
	return rec, nil
}

func (s *Service) TriggerChecklist(ctx context.Context, businessID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error) {
	if businessID <= 0 || !hasOnboardingManagePermission(ctx, businessID) {
		return nil, domain.ErrChecklistAccessDenied
	}
	belongs, err := s.repo.EmployeeBelongsToBusiness(ctx, businessID, employeeID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, domain.ErrEmployeeNotFound
	}
	if _, err := s.repo.GetTemplate(ctx, businessID, templateID); err != nil {
		return nil, err
	}
	rec, err := s.repo.TriggerChecklist(ctx, businessID, employeeID, templateID, triggerDate)
	if err != nil {
		return nil, err
	}
	invalidateHROnboardingScope(context.WithoutCancel(ctx), businessID)
	return rec, nil
}

func (s *Service) CompleteChecklistItem(ctx context.Context, businessID, itemID, completedBy int64, attachments ...*string) error {
	if businessID <= 0 {
		return domain.ErrChecklistAccessDenied
	}
	var attachmentURL *string
	if len(attachments) > 0 {
		attachmentURL = attachments[0]
	}
	item, err := s.repo.GetChecklistItem(ctx, businessID, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return domain.ErrChecklistItemNotFound
	}
	if attachmentURL != nil {
		normalized := strings.TrimSpace(*attachmentURL)
		if err := domain.ValidateAttachmentURL(normalized); err != nil {
			return err
		}
		attachmentURL = &normalized
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID != completedBy {
		return domain.ErrChecklistAccessDenied
	}
	if !hasOnboardingManagePermission(ctx, businessID) {
		callerEmployeeID, err := s.repo.GetEmployeeIDByUserID(ctx, businessID, completedBy)
		if err != nil || callerEmployeeID <= 0 || item.AssignedToEmployeeID != callerEmployeeID {
			return domain.ErrChecklistAccessDenied
		}
	}
	if item.Status == domain.ItemStatusCompleted || item.Status == domain.ItemStatusCompletedLegacy {
		return domain.ErrChecklistAlreadyCompleted
	}

	// The repository performs the conditional item update, completion counts,
	// and parent status transition in one transaction. A later count/status
	// failure therefore rolls the item write back instead of returning 500
	// after a committed completion.
	if err := s.repo.CompleteChecklistItemWithAttachment(ctx, businessID, itemID, completedBy, attachmentURL); err != nil {
		return err
	}
	invalidateHROnboardingScope(context.WithoutCancel(ctx), businessID)
	return nil
}

func (s *Service) GetEmployeeChecklist(ctx context.Context, businessID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error) {
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("cklst|e=%d|c=%d", employeeID, checklistID))
	return cacher.Of[*domain.EmployeeChecklistRecord]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (*domain.EmployeeChecklistRecord, error) {
			return s.repo.GetEmployeeChecklist(ctx, businessID, employeeID, checklistID)
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
}

// GetLatestEmployeeChecklist resolves the employee's most recent checklist for
// routes that only carry an employee id (GET /employees/{employee_id}).
func (s *Service) GetLatestEmployeeChecklist(ctx context.Context, businessID, employeeID int64) (*domain.EmployeeChecklistRecord, error) {
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("cklst|e=%d|latest", employeeID))
	return cacher.Of[*domain.EmployeeChecklistRecord]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (*domain.EmployeeChecklistRecord, error) {
			return s.repo.GetLatestEmployeeChecklist(ctx, businessID, employeeID)
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
}

func (s *Service) ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, params domain.ListEmployeeChecklistsParams) ([]domain.EmployeeChecklistRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOnboardingScope, businessID, fmt.Sprintf("emp_cklsts|e=%d|l=%d|o=%d", employeeID, params.Limit, params.Offset))
	res, err := cacher.Of[onboardingListResult[domain.EmployeeChecklistRecord]]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (onboardingListResult[domain.EmployeeChecklistRecord], error) {
			items, total, err := s.repo.ListEmployeeChecklists(ctx, businessID, employeeID, params.Limit, params.Offset)
			return onboardingListResult[domain.EmployeeChecklistRecord]{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOnboardingTTL),
	)
	return res.Items, res.Total, err
}
