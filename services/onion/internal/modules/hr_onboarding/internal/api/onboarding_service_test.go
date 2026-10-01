package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

// memberCtx authenticates a member whose user id (1002) differs from the
// business id (1001), the case the endpoints used to confuse.
func memberCtx(ctx context.Context) context.Context {
	return authPkg.WithAuthUser(ctx, &authPkg.AuthUser{
		UserID: 1002, BusinessID: testutil.BusinessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: testutil.BusinessID, AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppHR, Licensed: true, HasAccess: true,
				Permissions: []string{"hr:employees:manage"},
			}},
		},
	})
}

func ensureMemberBusinessAccess(t *testing.T, db *sql.DB) {
	t.Helper()
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO authorization_business_members (business_id, user_id, member_type, status)
		 VALUES (?, ?, 'member', 'active')
		 ON DUPLICATE KEY UPDATE status = 'active'`,
		testutil.BusinessID, 1002,
	); err != nil {
		t.Fatalf("grant member business access: %v", err)
	}
}

func createTemplateWithItem(t *testing.T, db *sql.DB, repo *infrastructure.Repository, name string) *domain.ChecklistTemplateRecord {
	t.Helper()

	ctx := context.Background()
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        name,
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID); err != nil {
			t.Errorf("cleanup template: %v", err)
		}
	})
	if _, err := repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "API regression task",
		AssigneeRole:  "employee",
		DueOffsetDays: 1,
	}); err != nil {
		t.Fatalf("CreateTemplateItem: %v", err)
	}

	return tmpl
}

// TestOnboardingServiceGetEmployeeChecklistReturnsEmployeeChecklist covers the
// route GET /hr/onboarding/v1/employees/{employee_id}, which carries no
// checklist id: the handler used to pass the employee id as the checklist id,
// so it only returned data when a checklist primary key happened to equal the
// employee id.
func TestOnboardingServiceGetEmployeeChecklistReturnsEmployeeChecklist(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureMemberBusinessAccess(t, db)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	service := application.NewService(ctx, db)

	tmpl := createTemplateWithItem(t, db, repo, "API Get Checklist Template")

	const employeeID int64 = 4111
	checklist, err := repo.TriggerChecklist(ctx, testutil.OwnerID, employeeID, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist: %v", err)
	}
	if checklist.ID == employeeID {
		t.Skipf("fixture collision: checklist id %d equals employee id; rerun to get a fresh id", checklist.ID)
	}

	resp, err := NewOnboardingService(service).GetEmployeeChecklist(memberCtx(ctx), &hronboardingv1.GetEmployeeChecklistRequest{
		EmployeeId: employeeID,
	})
	if err != nil {
		t.Fatalf("GetEmployeeChecklist: %v", err)
	}
	if resp.GetChecklist().GetId() != checklist.ID {
		t.Errorf("checklist id = %d, want %d", resp.GetChecklist().GetId(), checklist.ID)
	}
	if resp.GetChecklist().GetEmployeeId() != employeeID {
		t.Errorf("employee id = %d, want %d", resp.GetChecklist().GetEmployeeId(), employeeID)
	}
	if len(resp.GetChecklist().GetItems()) == 0 {
		t.Error("expected the checklist items copied from the template")
	}
}

// TestOnboardingServiceUpdateChecklistTemplateResolvesBusinessID covers the
// PUT /hr/onboarding/v1/templates/{role_id} handler, which scoped the update by
// the caller's user id instead of the resolved business id: a member could then
// modify another tenant's template whenever users.id collided with
// businesses.id, and in the normal case their own tenant's template 404ed.
func TestOnboardingServiceUpdateChecklistTemplateResolvesBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureMemberBusinessAccess(t, db)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	service := application.NewService(ctx, db)

	tmpl := createTemplateWithItem(t, db, repo, "API Update Template")

	_, err := NewOnboardingService(service).UpdateChecklistTemplate(memberCtx(ctx), &hronboardingv1.UpdateChecklistTemplateRequest{
		RoleId: tmpl.ID,
		Body:   &hronboardingv1.UpdateChecklistTemplateBody{Name: "Updated By Member"},
	})
	if err != nil {
		t.Fatalf("UpdateChecklistTemplate: %v (status %d)", err, oneKitStatus(err))
	}

	var name, triggerType string
	if err := db.QueryRowContext(ctx, "SELECT name, trigger_type FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID).Scan(&name, &triggerType); err != nil {
		t.Fatalf("read template after update: %v", err)
	}
	if name != "Updated By Member" {
		t.Errorf("template name = %q, want %q", name, "Updated By Member")
	}
	if triggerType != "hire" {
		t.Errorf("template trigger_type = %q, want preserved value %q", triggerType, "hire")
	}
}

func TestUpdateChecklistTemplateMissingTemplateReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureMemberBusinessAccess(t, db)
	ctx := context.Background()
	service := application.NewService(ctx, db)

	_, err := NewOnboardingService(service).UpdateChecklistTemplate(memberCtx(ctx), &hronboardingv1.UpdateChecklistTemplateRequest{
		RoleId: 999999999,
		Body:   &hronboardingv1.UpdateChecklistTemplateBody{Name: "Missing Template"},
	})
	if oneKitStatus(err) != http.StatusNotFound {
		t.Errorf("expected 404, got %v", err)
	}
}
