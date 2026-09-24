package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/testutil"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func authContext(userID, businessID int64, permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID: userID, BusinessID: businessID, Role: "member",
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: businessID, AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppHR, Licensed: true, HasAccess: true,
				Permissions: permissions,
				Scopes:      []authorization.Scope{{Type: "department", ID: "1"}},
			}},
		},
	})
}

func uniqueBusinessID() int64 {
	return 8_700_000_000 + time.Now().UnixNano()%100_000_000
}

func ensureRegressionBusinessAccess(t *testing.T, db *sql.DB, businessID int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`INSERT IGNORE INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Onboarding regression business %d", businessID), testutil.OwnerID,
	); err != nil {
		t.Fatalf("create regression business: %v", err)
	}
	if _, err := db.ExecContext(context.Background(),
		`INSERT IGNORE INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'member')`,
		1002, businessID,
	); err != nil {
		t.Fatalf("grant regression business access: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?", 1002, businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
	})
}

func createTemplateForRegression(t *testing.T, db *sql.DB, repo *infrastructure.Repository, businessID int64, name, triggerType string) *domain.ChecklistTemplateRecord {
	t.Helper()
	tmpl, err := repo.CreateTemplate(context.Background(), businessID, domain.CreateTemplateParams{
		Name: name, TriggerType: triggerType,
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	return tmpl
}

func createEmployeeForRegression(t *testing.T, db *sql.DB, businessID, employeeID int64, userID *int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_employees (id, owner_id, business_id, user_id, first_name, last_name, email, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'active')`,
		employeeID, businessID, businessID, userID,
		"Regression", "Employee", fmt.Sprintf("onboarding-%d@example.invalid", employeeID),
	)
	if err != nil {
		t.Fatalf("create employee %d: %v", employeeID, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_employees WHERE id = ?", employeeID)
	})
}

func withItemID(ctx context.Context, id int64) context.Context {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", strconv.FormatInt(id, 10))
	return context.WithValue(ctx, chi.RouteCtxKey, routeCtx)
}

func TestOnboardingServiceListTemplatesUsesBusinessAndTriggerFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	businessID := uniqueBusinessID()
	ensureRegressionBusinessAccess(t, db, businessID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	nameSuffix := time.Now().UnixNano()

	_ = createTemplateForRegression(t, db, repo, businessID, fmt.Sprintf("Regression hire %d", nameSuffix), "hire")
	_ = createTemplateForRegression(t, db, repo, businessID, fmt.Sprintf("Regression offboarding %d", nameSuffix), "offboarding")
	foreignBusinessID := uniqueBusinessID()
	_ = createTemplateForRegression(t, db, repo, foreignBusinessID, fmt.Sprintf("Regression foreign %d", nameSuffix), "hire")

	// Avoid a stale entry from another test using the same production-like key.
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix("hr_onboarding", businessID))

	trigger := "hire"
	res, err := NewOnboardingService(application.NewService(ctx, db)).ListChecklistTemplates(
		authContext(1002, businessID),
		&hronboardingv1.ListChecklistTemplatesRequest{TriggerType: &trigger},
	)
	if err != nil {
		t.Fatalf("ListChecklistTemplates: %v", err)
	}
	if res.GetTotal() < 1 || len(res.GetTemplates()) < 1 {
		t.Fatalf("filtered templates = %d (total %d), want at least the hire template", len(res.GetTemplates()), res.GetTotal())
	}
	wantName := fmt.Sprintf("Regression hire %d", nameSuffix)
	foundHire := false
	for _, tmpl := range res.GetTemplates() {
		if tmpl.GetName() == wantName {
			foundHire = true
		}
		if tmpl.GetName() == fmt.Sprintf("Regression offboarding %d", nameSuffix) {
			t.Fatalf("offboarding template leaked through hire filter: %q", tmpl.GetName())
		}
	}
	if !foundHire {
		t.Fatalf("filtered response did not contain %q: %+v", wantName, res.GetTemplates())
	}
}

func TestOnboardingServiceTriggerRejectsForeignEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	businessID := uniqueBusinessID()
	ensureRegressionBusinessAccess(t, db, businessID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	tmpl := createTemplateForRegression(t, db, repo, businessID, fmt.Sprintf("Regression trigger %d", time.Now().UnixNano()), "hire")

	employeeID := 8_800_000_000 + time.Now().UnixNano()%100_000_000
	createEmployeeForRegression(t, db, businessID, employeeID, nil)
	foreignEmployeeID := employeeID + 1
	createEmployeeForRegression(t, db, uniqueBusinessID(), foreignEmployeeID, nil)

	res, err := NewOnboardingService(application.NewService(ctx, db)).TriggerChecklist(
		authContext(1002, businessID, "hr:employees:manage"),
		&hronboardingv1.TriggerChecklistRequest{
			EmployeeId: employeeID, TemplateId: &tmpl.ID, TriggerDate: "2026-09-24",
		},
	)
	if err != nil {
		t.Fatalf("valid trigger: %v", err)
	}
	if res.GetChecklist().GetEmployeeId() != employeeID {
		t.Fatalf("checklist employee = %d, want %d", res.GetChecklist().GetEmployeeId(), employeeID)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_employee_checklist_items WHERE employee_checklist_id = ?", res.GetChecklist().GetId())
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_employee_checklists WHERE id = ?", res.GetChecklist().GetId())
	})

	_, err = NewOnboardingService(application.NewService(ctx, db)).TriggerChecklist(
		authContext(1002, businessID, "hr:employees:manage"),
		&hronboardingv1.TriggerChecklistRequest{
			EmployeeId: foreignEmployeeID, TemplateId: &tmpl.ID, TriggerDate: "2026-09-24",
		},
	)
	if oneKitStatus(err) != http.StatusNotFound {
		t.Fatalf("foreign employee trigger status = %d, want 404 (err=%v)", oneKitStatus(err), err)
	}
}

func TestCompleteChecklistItemPersistsCamelCaseAttachmentAndConflicts(t *testing.T) {
	db := testutil.OpenDB(t)
	businessID := uniqueBusinessID()
	ensureRegressionBusinessAccess(t, db, businessID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	tmpl := createTemplateForRegression(t, db, repo, businessID, fmt.Sprintf("Regression completion %d", time.Now().UnixNano()), "hire")
	if _, err := repo.CreateTemplateItem(ctx, businessID, tmpl.ID, domain.CreateTemplateItemParams{
		Title: "Upload proof", AssigneeRole: "employee",
	}); err != nil {
		t.Fatalf("create template item: %v", err)
	}
	employeeID := 8_900_000_000 + time.Now().UnixNano()%100_000_000
	checklist, err := repo.TriggerChecklist(ctx, businessID, employeeID, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("trigger checklist: %v", err)
	}
	if len(checklist.Items) != 1 {
		t.Fatalf("checklist items = %d, want 1", len(checklist.Items))
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_employee_checklist_items WHERE employee_checklist_id = ?", checklist.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_employee_checklists WHERE id = ?", checklist.ID)
	})
	itemID := checklist.Items[0].ID

	var req CompleteChecklistItemRequest
	if err := json.Unmarshal([]byte(`{"attachmentUrl":"https://cdn.example.test/proof.pdf"}`), &req); err != nil {
		t.Fatalf("decode completion request: %v", err)
	}
	if req.AttachmentURL == nil {
		t.Fatal("camelCase attachmentUrl was discarded")
	}
	handler := &CompleteChecklistItemHandler{Service: application.NewService(ctx, db)}
	if err := handler.Validate(ctx, &req); err != nil {
		t.Fatalf("validate completion request: %v", err)
	}
	completionCtx := withItemID(authContext(1002, businessID, "hr:employees:manage"), itemID)
	if _, err := handler.Serve(completionCtx, &req); err != nil {
		t.Fatalf("complete checklist item: %v", err)
	}

	var attachment, status string
	var completedBy int64
	if err := db.QueryRowContext(ctx,
		"SELECT COALESCE(attachment_url, ''), status, completed_by FROM hr_onboarding_employee_checklist_items WHERE id = ?",
		itemID,
	).Scan(&attachment, &status, &completedBy); err != nil {
		t.Fatalf("read completed item: %v", err)
	}
	if attachment != "https://cdn.example.test/proof.pdf" || status != "complete" || completedBy != 1002 {
		t.Fatalf("persisted item = attachment %q status %q completed_by %d", attachment, status, completedBy)
	}

	_, err = handler.Serve(completionCtx, &req)
	if oneKitStatus(err) != http.StatusConflict {
		t.Fatalf("second completion status = %d, want 409 (err=%v)", oneKitStatus(err), err)
	}
}

func TestOnboardingMutationsRequireManagePermission(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureMemberBusinessAccess(t, db)
	ctx := context.Background()
	service := application.NewService(ctx, db)
	client := NewOnboardingService(service)
	viewerCtx := authContext(1002, testutil.BusinessID, "hr:employees:view")

	_, err := client.CreateChecklistTemplate(viewerCtx, &hronboardingv1.CreateChecklistTemplateRequest{Name: "Denied"})
	if oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("viewer create status = %d, want 403", oneKitStatus(err))
	}

	_, err = client.UpdateChecklistTemplate(viewerCtx, &hronboardingv1.UpdateChecklistTemplateRequest{
		RoleId: 1, Body: &hronboardingv1.UpdateChecklistTemplateBody{Name: "Denied"},
	})
	if oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("viewer update status = %d, want 403", oneKitStatus(err))
	}
}
