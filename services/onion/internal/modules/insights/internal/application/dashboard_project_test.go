package application

import (
	"context"
	"errors"
	"testing"
	"time"

	financecontract "github.com/voxire/onion/internal/contracts/finance"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
)

// projectFlowFake answers for one visible project. Every other project is
// reported as not found, like Flow does for a private project the viewer is
// not a member of.
type projectFlowFake struct {
	visible  domain.DashboardProject
	viewer   int64
	progress domain.FlowProjectProgress
	time     domain.FlowProjectTimeTotals
	windows  []dateRange
	reads    int
	err      error
}

func (f *projectFlowFake) FlowDashboardProjects(_ context.Context, _, userID int64) ([]domain.DashboardProject, error) {
	if userID != f.viewer {
		return nil, f.err
	}
	return []domain.DashboardProject{f.visible}, f.err
}

func (f *projectFlowFake) FlowDashboardProject(_ context.Context, _, userID, projectID int64) (domain.DashboardProject, bool, error) {
	if f.err != nil {
		return domain.DashboardProject{}, false, f.err
	}
	if userID != f.viewer || projectID != f.visible.ID {
		return domain.DashboardProject{}, false, nil
	}
	return f.visible, true, nil
}

func (f *projectFlowFake) FlowProjectProgress(context.Context, int64, int64) (domain.FlowProjectProgress, error) {
	f.reads++
	return f.progress, nil
}

func (f *projectFlowFake) FlowProjectTime(_ context.Context, _, _ int64, thisFrom, thisTo, priorFrom, priorTo time.Time) (domain.FlowProjectTimeTotals, error) {
	f.reads++
	f.windows = append(f.windows, dateRange{thisFrom, thisTo}, dateRange{priorFrom, priorTo})
	return f.time, nil
}

type projectFinanceFake struct {
	totals financecontract.ProjectInvoiceTotals
	calls  int
}

func (f *projectFinanceFake) InsightsProjectInvoices(context.Context, int64, int64) (financecontract.ProjectInvoiceTotals, error) {
	f.calls++
	return f.totals, nil
}

func newProjectFakes() (*projectFlowFake, *projectFinanceFake) {
	flow := &projectFlowFake{
		visible: domain.DashboardProject{ID: 8, Name: "Website", Status: "active", Visibility: "private"},
		viewer:  7,
		progress: domain.FlowProjectProgress{
			TotalTasks: 8, CompletedTasks: 2, OverdueTasks: 3,
			Overdue: []domain.ProjectOverdueTask{{ID: 51, Title: "Fix footer", DueDate: "2026-10-01", AssigneeID: 7}},
		},
		time: domain.FlowProjectTimeTotals{AllTimeMinutes: 9_000, ThisMinutes: 2_400, PriorMinutes: 1_600},
	}
	finance := &projectFinanceFake{totals: financecontract.ProjectInvoiceTotals{
		InvoiceCount: 4, Invoiced: 200_000, Paid: 120_000, Outstanding: 80_000, Overdue: 30_000, OverdueCount: 1,
	}}
	return flow, finance
}

func TestReadProjectDashboard(t *testing.T) {
	flow, finance := newProjectFakes()
	service := NewService(nil, DashboardReaders{FlowProjects: flow, FinanceProjects: finance})

	got, err := service.readProjectDashboard(context.Background(), 42, 7, 8, true, hoursNow)
	if err != nil {
		t.Fatalf("readProjectDashboard: %v", err)
	}
	if got.Project != flow.visible {
		t.Errorf("Project = %+v, want %+v", got.Project, flow.visible)
	}
	if got.TotalTasks != 8 || got.CompletedTasks != 2 || got.ProgressPct != 25 {
		t.Errorf("progress = %d/%d (%v%%), want 2/8 (25%%)", got.CompletedTasks, got.TotalTasks, got.ProgressPct)
	}
	if got.OverdueTasks != 3 || len(got.Overdue) != 1 || got.Overdue[0].ID != 51 {
		t.Errorf("overdue = %d %+v, want 3 with task 51 listed", got.OverdueTasks, got.Overdue)
	}
	// Same month-to-date rule as the overview's hours card.
	if got.Logged.Value != 2_400 || got.Logged.DeltaPct != 50 || got.PriorMinutes != 1_600 || got.AllTimeMinutes != 9_000 {
		t.Errorf("hours = %+v, prior %d, all time %d; want 2400 up 50%% on 1600, 9000 all time", got.Logged, got.PriorMinutes, got.AllTimeMinutes)
	}
	if len(flow.windows) != 2 || flow.windows[0] != hoursThisMTD || flow.windows[1] != hoursPriorMTD {
		t.Errorf("windows = %#v, want %#v and %#v", flow.windows, hoursThisMTD, hoursPriorMTD)
	}
	want := domain.ProjectInvoices{InvoiceCount: 4, Invoiced: 200_000, Paid: 120_000, Outstanding: 80_000, Overdue: 30_000, OverdueCount: 1}
	if got.Invoices == nil || *got.Invoices != want {
		t.Errorf("Invoices = %+v, want %+v", got.Invoices, want)
	}
}

// A viewer without finance invoice access gets the dashboard without money,
// and finance is not even asked.
func TestReadProjectDashboardWithoutInvoices(t *testing.T) {
	flow, finance := newProjectFakes()
	service := NewService(nil, DashboardReaders{FlowProjects: flow, FinanceProjects: finance})

	got, err := service.readProjectDashboard(context.Background(), 42, 7, 8, false, hoursNow)
	if err != nil {
		t.Fatalf("readProjectDashboard: %v", err)
	}
	if got.Invoices != nil || finance.calls != 0 {
		t.Errorf("Invoices = %+v after %d finance reads, want none", got.Invoices, finance.calls)
	}
	if got.TotalTasks != 8 {
		t.Errorf("TotalTasks = %d, want the rest of the dashboard", got.TotalTasks)
	}
}

// A project the viewer may not see is not found, and nothing about it is read.
func TestReadProjectDashboardHidesProjectsTheViewerCannotSee(t *testing.T) {
	for _, tc := range []struct {
		name              string
		viewer, projectID int64
	}{
		{"private project, not a member", 99, 8},
		{"missing project", 7, 404},
		{"no project id", 7, 0},
		{"no viewer", 0, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flow, finance := newProjectFakes()
			service := NewService(nil, DashboardReaders{FlowProjects: flow, FinanceProjects: finance})
			_, err := service.readProjectDashboard(context.Background(), 42, tc.viewer, tc.projectID, true, hoursNow)
			if !errors.Is(err, ErrProjectNotFound) {
				t.Fatalf("err = %v, want ErrProjectNotFound", err)
			}
			if flow.reads != 0 || finance.calls != 0 {
				t.Errorf("read %d Flow figures and %d invoice totals for a hidden project", flow.reads, finance.calls)
			}
		})
	}
}

func TestReadProjectDashboardWithoutTasks(t *testing.T) {
	flow, finance := newProjectFakes()
	flow.progress = domain.FlowProjectProgress{}
	service := NewService(nil, DashboardReaders{FlowProjects: flow, FinanceProjects: finance})

	got, err := service.readProjectDashboard(context.Background(), 42, 7, 8, false, hoursNow)
	if err != nil {
		t.Fatalf("readProjectDashboard: %v", err)
	}
	if got.ProgressPct != 0 || got.Overdue == nil || len(got.Overdue) != 0 {
		t.Errorf("empty project = %v%% with overdue %#v, want 0%% and an empty list", got.ProgressPct, got.Overdue)
	}
}

func TestReadProjectDashboardFailsClosedWithoutReaders(t *testing.T) {
	flow, _ := newProjectFakes()
	if _, err := NewService(nil, DashboardReaders{}).readProjectDashboard(context.Background(), 42, 7, 8, false, hoursNow); !errors.Is(err, errDashboardReadersUnavailable) {
		t.Errorf("no Flow reader: err = %v", err)
	}
	if _, err := NewService(nil, DashboardReaders{FlowProjects: flow}).readProjectDashboard(context.Background(), 42, 7, 8, true, hoursNow); !errors.Is(err, errDashboardReadersUnavailable) {
		t.Errorf("invoices without a finance reader: err = %v", err)
	}
	if _, err := NewService(nil, DashboardReaders{}).ListDashboardProjects(context.Background(), 42, 7); !errors.Is(err, errDashboardReadersUnavailable) {
		t.Errorf("list without a Flow reader: err = %v", err)
	}
}

func TestListDashboardProjects(t *testing.T) {
	flow, _ := newProjectFakes()
	service := NewService(nil, DashboardReaders{FlowProjects: flow})
	got, err := service.ListDashboardProjects(context.Background(), 42, 7)
	if err != nil || len(got) != 1 || got[0] != flow.visible {
		t.Errorf("projects = %+v, %v; want the visible project", got, err)
	}
	none, err := service.ListDashboardProjects(context.Background(), 42, 99)
	if err != nil || none == nil || len(none) != 0 {
		t.Errorf("projects for an outsider = %#v, %v; want an empty list", none, err)
	}
}
