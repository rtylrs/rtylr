package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

type mockRepo struct {
	template         *domain.ChecklistTemplateRecord
	templates        []domain.ChecklistTemplateRecord
	totalTmpl        int64
	item             *domain.ChecklistItemRecord
	checklistItem    *domain.EmployeeChecklistItemRecord
	checklist        *domain.EmployeeChecklistRecord
	checklists       []domain.EmployeeChecklistRecord
	totalCklst       int64
	totalItems       int
	completedItems   int
	employeeBelongs  bool
	employeeErr      error
	callerEmployeeID int64
	completeErr      error
	completeCalls    int
	attachmentURL    *string
	statsErr         error
	listParams       domain.ListTemplatesParams
	updateParams     domain.UpdateTemplateParams
	err              error
}

func serviceContext(userID int64, permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID: userID, BusinessID: 1,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: 1, AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppHR, Licensed: true, HasAccess: true,
				Permissions: permissions,
			}},
		},
	})
}

func (m *mockRepo) CreateTemplate(_ context.Context, _ int64, _ domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	return m.template, m.err
}

func (m *mockRepo) GetTemplate(_ context.Context, _, _ int64) (*domain.ChecklistTemplateRecord, error) {
	return m.template, m.err
}

func (m *mockRepo) ListTemplates(_ context.Context, _ int64, _, _ int32) ([]domain.ChecklistTemplateRecord, int64, error) {
	return m.templates, m.totalTmpl, m.err
}

func (m *mockRepo) ListTemplatesFiltered(_ context.Context, _ int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error) {
	m.listParams = params
	return m.templates, m.totalTmpl, m.err
}

func (m *mockRepo) UpdateTemplate(_ context.Context, _, _ int64, params domain.UpdateTemplateParams) error {
	m.updateParams = params
	return m.err
}

func (m *mockRepo) CreateTemplateItem(_ context.Context, _, _ int64, _ domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	return m.item, m.err
}

func (m *mockRepo) ListTemplateItems(_ context.Context, _, _ int64) ([]domain.ChecklistItemRecord, error) {
	return nil, m.err
}

func (m *mockRepo) TriggerChecklist(_ context.Context, _, _, _ int64, _ time.Time) (*domain.EmployeeChecklistRecord, error) {
	return m.checklist, m.err
}

func (m *mockRepo) GetEmployeeChecklist(_ context.Context, _, _, _ int64) (*domain.EmployeeChecklistRecord, error) {
	return m.checklist, m.err
}

func (m *mockRepo) GetLatestEmployeeChecklist(_ context.Context, _, _ int64) (*domain.EmployeeChecklistRecord, error) {
	return m.checklist, m.err
}

func (m *mockRepo) ListEmployeeChecklists(_ context.Context, _, _ int64, _, _ int32) ([]domain.EmployeeChecklistRecord, int64, error) {
	return m.checklists, m.totalCklst, m.err
}

func (m *mockRepo) UpdateEmployeeChecklistStatus(_ context.Context, _, _ int64, _ string) error {
	return m.err
}

func (m *mockRepo) CreateEmployeeChecklistItem(_ context.Context, _, _, _, _ int64, _ string) error {
	return m.err
}

func (m *mockRepo) GetChecklistItem(_ context.Context, _, _ int64) (*domain.EmployeeChecklistItemRecord, error) {
	return m.checklistItem, m.err
}
func (m *mockRepo) CompleteChecklistItem(_ context.Context, _, _, _ int64) error {
	return m.CompleteChecklistItemWithAttachment(context.Background(), 0, 0, 0, nil)
}

func (m *mockRepo) CompleteChecklistItemWithAttachment(_ context.Context, _, _, _ int64, attachmentURL *string) error {
	m.completeCalls++
	m.attachmentURL = attachmentURL
	return m.completeErr
}

func (m *mockRepo) GetChecklistCompletionStats(_ context.Context, _, _ int64) (int, int, error) {
	return m.totalItems, m.completedItems, m.statsErr
}

func (m *mockRepo) EmployeeBelongsToBusiness(_ context.Context, _, _ int64) (bool, error) {
	if m.employeeErr != nil {
		return false, m.employeeErr
	}
	return m.employeeBelongs, nil
}

func (m *mockRepo) GetEmployeeIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	if m.employeeErr != nil {
		return 0, m.employeeErr
	}
	return m.callerEmployeeID, nil
}

func (m *mockRepo) ListEmployeeChecklistItems(_ context.Context, _, _ int64) ([]domain.EmployeeChecklistItemRecord, error) {
	return nil, m.err
}

// ---- ListTemplates ----

func TestListTemplates_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.ChecklistTemplateRecord{{ID: 1, Name: "Onboarding"}}
	s := &Service{repo: &mockRepo{templates: items, totalTmpl: 1}}
	got, total, err := s.ListTemplates(context.Background(), 1, domain.ListTemplatesParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || total != 1 {
		t.Errorf("expected 1 item, got %d (total=%d)", len(got), total)
	}
}

func TestListTemplates_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, _, err := s.ListTemplates(context.Background(), 1, domain.ListTemplatesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %+v", got)
	}
}

func TestListTemplates_PropagatesRepoError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.ListTemplates(context.Background(), 1, domain.ListTemplatesParams{Limit: 10})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListTemplates_DefaultsLimitWhenZero(t *testing.T) {
	s := &Service{repo: &mockRepo{templates: []domain.ChecklistTemplateRecord{{ID: 1}}}}
	// Limit=0 should default to 20 and succeed
	_, _, err := s.ListTemplates(context.Background(), 1, domain.ListTemplatesParams{Limit: 0})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---- CreateTemplate ----

func TestCreateTemplate_ReturnsRecord(t *testing.T) {
	rec := &domain.ChecklistTemplateRecord{ID: 10, Name: "New Hire"}
	s := &Service{repo: &mockRepo{template: rec}}
	got, err := s.CreateTemplate(serviceContext(1, "hr:employees:manage"), 1, domain.CreateTemplateParams{Name: "New Hire"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 10 {
		t.Errorf("expected ID=10, got %d", got.ID)
	}
}

func TestCreateTemplate_DefaultsTriggerType(t *testing.T) {
	called := false
	repo := &captureCreateTemplateRepo{onCreate: func(p domain.CreateTemplateParams) {
		called = true
		if p.TriggerType != "onboarding" {
			panic("expected onboarding trigger type")
		}
	}}
	s := &Service{repo: repo}
	_, _ = s.CreateTemplate(serviceContext(1, "hr:employees:manage"), 1, domain.CreateTemplateParams{Name: "Test"})
	if !called {
		t.Error("expected CreateTemplate to be called")
	}
}

func TestCreateTemplate_PropagatesRepoError(t *testing.T) {
	want := errors.New("create error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateTemplate(serviceContext(1, "hr:employees:manage"), 1, domain.CreateTemplateParams{Name: "Test"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// captureCreateTemplateRepo is a specialized mock to capture CreateTemplate params.
type captureCreateTemplateRepo struct {
	mockRepo
	onCreate func(domain.CreateTemplateParams)
}

func (m *captureCreateTemplateRepo) CreateTemplate(_ context.Context, _ int64, p domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	if m.onCreate != nil {
		m.onCreate(p)
	}
	return &domain.ChecklistTemplateRecord{}, nil
}

// ---- UpdateTemplate ----

func TestUpdateTemplate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{template: &domain.ChecklistTemplateRecord{ID: 1}}}
	err := s.UpdateTemplate(serviceContext(1, "hr:employees:manage"), 1, 1, domain.UpdateTemplateParams{Name: "Updated"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateTemplate_PropagatesGetTemplateError(t *testing.T) {
	want := domain.ErrTemplateNotFound
	s := &Service{repo: &mockRepo{err: want}}
	err := s.UpdateTemplate(serviceContext(1, "hr:employees:manage"), 1, 99, domain.UpdateTemplateParams{})
	if !errors.Is(err, want) {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

// ---- CompleteChecklistItem ----

func TestCompleteChecklistItem_AlreadyCompletedReturnsError(t *testing.T) {
	item := &domain.EmployeeChecklistItemRecord{ID: 1, Status: "complete"}
	s := &Service{repo: &mockRepo{checklistItem: item}}
	err := s.CompleteChecklistItem(serviceContext(99, "hr:employees:manage"), 1, 1, 99)
	if !errors.Is(err, domain.ErrChecklistAlreadyCompleted) {
		t.Errorf("expected ErrChecklistAlreadyCompleted, got %v", err)
	}
}

func TestCompleteChecklistItem_PropagatesGetItemError(t *testing.T) {
	want := domain.ErrChecklistItemNotFound
	s := &Service{repo: &mockRepo{err: want}}
	err := s.CompleteChecklistItem(context.Background(), 1, 99, 1)
	if !errors.Is(err, want) {
		t.Errorf("expected ErrChecklistItemNotFound, got %v", err)
	}
}

func TestCompleteChecklistItem_AllowsAssignedEmployeeAndForwardsAttachment(t *testing.T) {
	attachment := "https://cdn.example.test/checklist.pdf"
	item := &domain.EmployeeChecklistItemRecord{
		ID: 1, EmployeeChecklistID: 7, AssignedToEmployeeID: 42, Status: "incomplete",
	}
	repo := &mockRepo{checklistItem: item, callerEmployeeID: 42}
	s := &Service{repo: repo}

	if err := s.CompleteChecklistItem(serviceContext(99), 1, 1, 99, &attachment); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.completeCalls != 1 || repo.attachmentURL == nil || *repo.attachmentURL != attachment {
		t.Fatalf("completion call = %d, attachment = %v", repo.completeCalls, repo.attachmentURL)
	}
}

func TestCompleteChecklistItem_RejectsAnotherEmployee(t *testing.T) {
	item := &domain.EmployeeChecklistItemRecord{
		ID: 1, EmployeeChecklistID: 7, AssignedToEmployeeID: 42, Status: "incomplete",
	}
	repo := &mockRepo{checklistItem: item, callerEmployeeID: 43}
	s := &Service{repo: repo}

	err := s.CompleteChecklistItem(serviceContext(99), 1, 1, 99)
	if !errors.Is(err, domain.ErrChecklistAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
	if repo.completeCalls != 0 {
		t.Fatal("unauthorized completion reached the repository")
	}
}

func TestCompleteChecklistItem_ManagePermissionAllowsManager(t *testing.T) {
	item := &domain.EmployeeChecklistItemRecord{ID: 1, EmployeeChecklistID: 7, Status: "incomplete"}
	repo := &mockRepo{checklistItem: item}
	s := &Service{repo: repo}

	if err := s.CompleteChecklistItem(serviceContext(99, "hr:employees:manage"), 1, 1, 99); err != nil {
		t.Fatalf("unexpected manager completion error: %v", err)
	}
	if repo.completeCalls != 1 {
		t.Fatalf("completion calls = %d, want 1", repo.completeCalls)
	}
}

func TestCompleteChecklistItem_DoesNotReturnPostWriteStatsFailure(t *testing.T) {
	item := &domain.EmployeeChecklistItemRecord{ID: 1, EmployeeChecklistID: 7, AssignedToEmployeeID: 42, Status: "incomplete"}
	repo := &mockRepo{
		checklistItem: item, callerEmployeeID: 42,
		statsErr: errors.New("stats query failed"),
	}
	s := &Service{repo: repo}

	if err := s.CompleteChecklistItem(serviceContext(99), 1, 1, 99); err != nil {
		t.Fatalf("atomic completion should not expose a post-write stats error: %v", err)
	}
	if repo.completeCalls != 1 {
		t.Fatalf("completion calls = %d, want 1", repo.completeCalls)
	}
}

func TestCreateTemplate_ViewerCannotCreate(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}

	_, err := s.CreateTemplate(serviceContext(99, "hr:employees:view"), 1, domain.CreateTemplateParams{Name: "Denied"})
	if !errors.Is(err, domain.ErrChecklistAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
	if repo.template != nil {
		t.Fatal("viewer creation reached the repository")
	}
}

func TestUpdateTemplate_PreservesOmittedTriggerType(t *testing.T) {
	repo := &mockRepo{template: &domain.ChecklistTemplateRecord{ID: 1, TriggerType: "hire"}}
	s := &Service{repo: repo}

	if err := s.UpdateTemplate(serviceContext(1, "hr:employees:manage"), 1, 1, domain.UpdateTemplateParams{Name: "Renamed"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updateParams.TriggerType == nil || *repo.updateParams.TriggerType != "hire" {
		t.Fatalf("trigger patch = %v, want preserved hire", repo.updateParams.TriggerType)
	}
}

func TestTriggerChecklist_RejectsForeignEmployee(t *testing.T) {
	repo := &mockRepo{template: &domain.ChecklistTemplateRecord{ID: 1}, employeeBelongs: false}
	s := &Service{repo: repo}

	_, err := s.TriggerChecklist(serviceContext(1, "hr:employees:manage"), 1, 99, 1, time.Now())
	if !errors.Is(err, domain.ErrEmployeeNotFound) {
		t.Fatalf("expected employee not found, got %v", err)
	}
}

func TestListTemplates_PreservesTriggerFilter(t *testing.T) {
	trigger := "hire"
	repo := &mockRepo{}
	s := &Service{repo: repo}

	if _, _, err := s.ListTemplates(context.Background(), 1, domain.ListTemplatesParams{TriggerType: &trigger}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listParams.TriggerType == nil || *repo.listParams.TriggerType != trigger {
		t.Fatalf("trigger filter = %v, want %q", repo.listParams.TriggerType, trigger)
	}
}

// ---- ListEmployeeChecklists ----

func TestListEmployeeChecklists_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.EmployeeChecklistRecord{{ID: 1, EmployeeID: 5}}
	s := &Service{repo: &mockRepo{checklists: items, totalCklst: 1}}
	got, total, err := s.ListEmployeeChecklists(context.Background(), 1, 5, domain.ListEmployeeChecklistsParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || total != 1 {
		t.Errorf("expected 1 item, got %d (total=%d)", len(got), total)
	}
}

func TestListEmployeeChecklists_PropagatesRepoError(t *testing.T) {
	want := errors.New("list error")
	s := &Service{repo: &mockRepo{err: want}}
	_, _, err := s.ListEmployeeChecklists(context.Background(), 1, 5, domain.ListEmployeeChecklistsParams{Limit: 10})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- TriggerChecklist ----

func TestTriggerChecklist_HappyPath(t *testing.T) {
	tmpl := &domain.ChecklistTemplateRecord{ID: 1}
	cklst := &domain.EmployeeChecklistRecord{ID: 50}
	s := &Service{repo: &mockRepo{template: tmpl, checklist: cklst, employeeBelongs: true}}
	got, err := s.TriggerChecklist(serviceContext(1, "hr:employees:manage"), 1, 10, 1, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 50 {
		t.Errorf("expected ID=50, got %d", got.ID)
	}
}

func TestTriggerChecklist_PropagatesTemplateNotFoundError(t *testing.T) {
	want := domain.ErrTemplateNotFound
	s := &Service{repo: &mockRepo{err: want, employeeBelongs: true}}
	_, err := s.TriggerChecklist(serviceContext(1, "hr:employees:manage"), 1, 10, 99, time.Now())
	if !errors.Is(err, want) {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}
