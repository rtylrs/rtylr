package api

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/testutil"
	insightsdashboardv1 "github.com/voxire/voxie/gen/go/insights/dashboard/v1"
)

func TestToProjectDashboardResponse(t *testing.T) {
	got := toProjectDashboardResponse(domain.ProjectDashboard{
		Project:        domain.DashboardProject{ID: 8, Name: "Website", Status: "active", Visibility: "public"},
		TotalTasks:     8,
		CompletedTasks: 2,
		ProgressPct:    25,
		OverdueTasks:   3,
		Overdue: []domain.ProjectOverdueTask{
			{ID: 51, Title: "Fix footer", DueDate: "2026-10-01", AssigneeID: 7, AssigneeName: "Abed Amouneh"},
			{ID: 52, Title: "Write copy", DueDate: "2026-10-02"},
			{ID: 53, Title: "Plan launch", DueDate: "2026-10-03", AssigneeID: 9},
		},
		Logged:         domain.CountCard{Value: 2_400, DeltaPct: 50},
		PriorMinutes:   1_600,
		AllTimeMinutes: 9_000,
		Invoices:       &domain.ProjectInvoices{InvoiceCount: 4, Invoiced: 200_000, Paid: 120_000, Outstanding: 80_000, Overdue: 30_000, OverdueCount: 1},
	})
	if got.Project.GetProjectId() != 8 || got.Project.GetProjectName() != "Website" || got.Project.GetVisibility() != "public" {
		t.Errorf("Project = %+v", got.Project)
	}
	if got.Progress.GetTotalTasks() != 8 || got.Progress.GetCompletedTasks() != 2 || got.Progress.GetPercent() != 25 {
		t.Errorf("Progress = %+v, want 2/8 at 25%%", got.Progress)
	}
	if got.OverdueTaskCount != 3 || len(got.OverdueTasks) != 3 {
		t.Fatalf("overdue = %d %+v", got.OverdueTaskCount, got.OverdueTasks)
	}
	if got.OverdueTasks[0].AssigneeId == nil || *got.OverdueTasks[0].AssigneeId != 7 {
		t.Errorf("assigned task = %+v, want assignee 7", got.OverdueTasks[0])
	}
	if got.OverdueTasks[0].AssigneeName == nil || *got.OverdueTasks[0].AssigneeName != "Abed Amouneh" {
		t.Errorf("assigned task = %+v, want assignee name Abed Amouneh", got.OverdueTasks[0])
	}
	// An unassigned task leaves assignee_id and assignee_name out rather than
	// sending 0 and "".
	if got.OverdueTasks[1].AssigneeId != nil || got.OverdueTasks[1].AssigneeName != nil {
		t.Errorf("unassigned task = %+v, want no assignee", got.OverdueTasks[1])
	}
	// An assignee whose name is unknown keeps the id and leaves the name out.
	if got.OverdueTasks[2].AssigneeId == nil || *got.OverdueTasks[2].AssigneeId != 9 || got.OverdueTasks[2].AssigneeName != nil {
		t.Errorf("unnamed assignee = %+v, want assignee 9 with no name", got.OverdueTasks[2])
	}
	if got.Hours.GetThisMonth().GetValue() != 2_400 || got.Hours.GetThisMonth().GetDeltaPct() != 50 || got.Hours.GetPriorMinutes() != 1_600 || got.Hours.GetAllTimeMinutes() != 9_000 {
		t.Errorf("Hours = %+v", got.Hours)
	}
	if got.Invoices == nil || got.Invoices.Invoiced != 200_000 || got.Invoices.Paid != 120_000 || got.Invoices.Outstanding != 80_000 || got.Invoices.Overdue != 30_000 || got.Invoices.OverdueCount != 1 || got.Invoices.InvoiceCount != 4 {
		t.Errorf("Invoices = %+v", got.Invoices)
	}
}

// Without finance access the invoices section is absent, not zero: a zero
// would read as "nothing invoiced". Cards and lists are never null.
func TestToProjectDashboardResponseWithoutInvoices(t *testing.T) {
	got := toProjectDashboardResponse(domain.ProjectDashboard{})
	if got.Invoices != nil {
		t.Errorf("Invoices = %+v, want absent", got.Invoices)
	}
	if got.Project == nil || got.Progress == nil || got.Hours == nil || got.Hours.ThisMonth == nil || got.OverdueTasks == nil {
		t.Errorf("response = %+v, want every section present", got)
	}
}

// projectScope resolves every caller to one business, standing in for the
// membership lookup the server does.
type projectScope struct{ businessID int64 }

func (s projectScope) ActiveBusinessOwner(context.Context, int64) (int64, error) {
	return s.businessID, nil
}

func (s projectScope) UserBusiness(context.Context, int64, int64) (int64, error) {
	return s.businessID, nil
}

func projectViewer(userID int64, financeInvoices bool) *authPkg.AuthUser {
	apps := []authorization.EffectiveAppAccess{{
		AppID: authorization.AppInsights, Licensed: true, HasAccess: true,
		Permissions: []string{"insights:dashboard:view"},
	}}
	if financeInvoices {
		apps = append(apps, authorization.EffectiveAppAccess{
			AppID: authorization.AppFinance, Licensed: true, HasAccess: true,
			Permissions: []string{"finance:invoices:view"},
		})
	}
	return &authPkg.AuthUser{UserID: userID, BusinessID: projectDashBiz, EffectiveAccess: &authorization.EffectiveAccess{Apps: apps}}
}

const (
	projectDashBiz, projectDashOther                      int64 = 956800001, 956800002
	projectDashCreator, projectDashMember, projectDashOut int64 = 956800011, 956800012, 956800013
)

// seedProjectDashboard creates a public and a private project in one business
// and a public project in another, each with one task and one invoice.
func seedProjectDashboard(t *testing.T, db *sql.DB) (public, private, foreign int64) {
	t.Helper()
	ctx := context.Background()
	remove := func() {
		for _, q := range []string{
			`DELETE FROM finance_invoices WHERE business_id IN (?, ?)`,
			`DELETE FROM flow_tasks WHERE business_id IN (?, ?)`,
			`DELETE pm FROM flow_project_members pm JOIN flow_projects p ON p.id = pm.project_id WHERE p.business_id IN (?, ?)`,
			`DELETE FROM flow_projects WHERE business_id IN (?, ?)`,
		} {
			if _, err := db.ExecContext(ctx, q, projectDashBiz, projectDashOther); err != nil {
				t.Fatalf("clear: %v", err)
			}
		}
	}
	remove()
	t.Cleanup(remove)

	project := func(businessID int64, name, visibility string) int64 {
		res, err := db.ExecContext(ctx, `INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, visibility, created_by)
			VALUES (0, ?, ?, ?, ?, ?)`, projectDashCreator, businessID, name, visibility, projectDashCreator)
		if err != nil {
			t.Fatalf("insert project: %v", err)
		}
		id, _ := res.LastInsertId()
		if _, err := db.ExecContext(ctx, `INSERT INTO flow_tasks (project_id, workspace_id, owner_id, business_id, title, status, reporter_id)
			VALUES (?, 0, ?, ?, 'Task', 'done', ?)`, id, projectDashCreator, businessID, projectDashCreator); err != nil {
			t.Fatalf("insert task: %v", err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO finance_invoices (owner_id, business_id, invoice_number, client_name, project_id, status, issue_date, due_date, total, amount_paid, amount_due)
			VALUES (?, ?, ?, 'Client', ?, 'sent', CURDATE(), CURDATE(), 250.00, 0, 250.00)`, businessID, businessID, name, id); err != nil {
			t.Fatalf("insert invoice: %v", err)
		}
		return id
	}
	public = project(projectDashBiz, "Public site", "public")
	private = project(projectDashBiz, "Private audit", "private")
	foreign = project(projectDashOther, "Other business", "public")
	if _, err := db.ExecContext(ctx, `INSERT INTO flow_project_members (project_id, user_id, role) VALUES (?, ?, 'member')`, private, projectDashMember); err != nil {
		t.Fatalf("insert member: %v", err)
	}
	return public, private, foreign
}

// TestProjectDashboardVisibility runs the endpoints against the real Flow and
// finance readers: a private project answers 404 to a non-member exactly like
// a missing project or another business's, and money needs finance access.
func TestProjectDashboardVisibility(t *testing.T) {
	db := testutil.OpenDB(t)
	t.Cleanup(tenant.SetScopeResolver(projectScope{businessID: projectDashBiz}))
	public, private, foreign := seedProjectDashboard(t, db)
	dashboardAPI := NewDashboardAPI(application.NewService(testRepository(db), testDashboardReaders(db)), nil)

	get := func(user *authPkg.AuthUser, projectID int64) (*insightsdashboardv1.GetProjectDashboardResponse, error) {
		return dashboardAPI.GetProjectDashboard(authPkg.WithAuthUser(context.Background(), user), &insightsdashboardv1.GetProjectDashboardRequest{ProjectId: projectID})
	}

	for _, tc := range []struct {
		name      string
		user      *authPkg.AuthUser
		projectID int64
	}{
		{"private project, not a member", projectViewer(projectDashOut, true), private},
		{"another business's project", projectViewer(projectDashCreator, true), foreign},
		{"missing project", projectViewer(projectDashCreator, true), foreign + 1000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := get(tc.user, tc.projectID)
			if insightsStatus(err) != http.StatusNotFound || resp != nil {
				t.Fatalf("got %+v, %v; want 404", resp, err)
			}
		})
	}

	t.Run("member opens the private project", func(t *testing.T) {
		resp, err := get(projectViewer(projectDashMember, false), private)
		if err != nil {
			t.Fatalf("GetProjectDashboard: %v", err)
		}
		if resp.Project.GetProjectId() != private || resp.Project.GetVisibility() != "private" || resp.Progress.GetCompletedTasks() != 1 || resp.Progress.GetPercent() != 100 {
			t.Errorf("response = %+v", resp)
		}
		// No finance access: no money, and no error either.
		if resp.Invoices != nil {
			t.Errorf("Invoices = %+v without finance access", resp.Invoices)
		}
	})

	t.Run("finance access adds the project's invoices", func(t *testing.T) {
		resp, err := get(projectViewer(projectDashOut, true), public)
		if err != nil {
			t.Fatalf("GetProjectDashboard: %v", err)
		}
		if resp.Invoices == nil || resp.Invoices.InvoiceCount != 1 || resp.Invoices.Invoiced != 25_000 || resp.Invoices.Outstanding != 25_000 || resp.Invoices.Overdue != 0 {
			t.Errorf("Invoices = %+v, want the one 250.00 invoice, not yet overdue", resp.Invoices)
		}
	})

	t.Run("the picker lists only what the caller may open", func(t *testing.T) {
		ids := func(user *authPkg.AuthUser) []int64 {
			resp, err := dashboardAPI.ListProjects(authPkg.WithAuthUser(context.Background(), user), &insightsdashboardv1.ListProjectsRequest{})
			if err != nil {
				t.Fatalf("ListProjects: %v", err)
			}
			out := []int64{}
			for _, p := range resp.Projects {
				out = append(out, p.GetProjectId())
			}
			return out
		}
		// Sorted by name: "Private audit" before "Public site".
		if got := ids(projectViewer(projectDashMember, false)); len(got) != 2 || got[0] != private || got[1] != public {
			t.Errorf("member sees %v, want [%d %d]", got, private, public)
		}
		if got := ids(projectViewer(projectDashOut, true)); len(got) != 1 || got[0] != public {
			t.Errorf("outsider sees %v, want [%d]", got, public)
		}
	})
}
