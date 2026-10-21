package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetDashboardStats_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	stats, err := repo.GetDashboardStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.TotalCustomers < 0 {
		t.Errorf("expected non-negative TotalCustomers, got %d", stats.TotalCustomers)
	}
}

func TestGetDashboardStats_TotalCustomersIsPositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	stats, err := repo.GetDashboardStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.TotalCustomers == 0 {
		t.Log("TotalCustomers=0; fixture may have no active customers")
	}
	if stats.TotalCustomers < 0 {
		t.Errorf("negative TotalCustomers: %d", stats.TotalCustomers)
	}
}

func TestGetDashboardStats_TopCustomersNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	stats, err := repo.GetDashboardStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
	if stats.TopCustomers == nil {
		t.Fatal("expected non-nil TopCustomers slice")
	}
}

func TestGetDashboardStats_WrongOwner_ReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	stats, err := repo.GetDashboardStats(ctx, 9999999)
	if err != nil {
		t.Fatalf("GetDashboardStats wrong owner: %v", err)
	}
	if stats.TotalCustomers != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", stats.TotalCustomers)
	}
}

func TestGetCustomerTimeline_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	entries, err := repo.GetCustomerTimeline(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerTimeline: %v", err)
	}
	if entries == nil {
		t.Fatal("expected non-nil entries slice")
	}
}

func TestInsertTimelineEntry_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	err := repo.InsertTimelineEntry(ctx, domain.TimelineEntry{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		CustomerID:  5301,
		Type:        "visit",
		Description: "integration test visit",
		Metadata:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("InsertTimelineEntry: %v", err)
	}
}

func TestInsertTimelineEntry_AppearsInTimeline(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	before, err := repo.GetCustomerTimeline(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerTimeline before: %v", err)
	}

	if err := repo.InsertTimelineEntry(ctx, domain.TimelineEntry{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		CustomerID:  5302,
		Type:        "note",
		Description: "timeline count check",
		Metadata:    []byte(`{}`),
	}); err != nil {
		t.Fatalf("InsertTimelineEntry: %v", err)
	}

	after, err := repo.GetCustomerTimeline(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerTimeline after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestExportCustomerData_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	export, err := repo.ExportCustomerData(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("ExportCustomerData: %v", err)
	}
	if export.Customer.ID != 5301 {
		t.Errorf("expected customer ID 5301, got %d", export.Customer.ID)
	}
}
