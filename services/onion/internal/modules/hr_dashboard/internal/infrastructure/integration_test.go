package infrastructure_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetWidgets_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	widgets, err := repo.GetWidgets(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	_ = widgets
}

func TestGetWidgets_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	widgets, err := repo.GetWidgets(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	for _, w := range widgets {
		if w.OwnerID != testutil.OwnerID {
			t.Errorf("widget %d belongs to wrong owner: %d", w.ID, w.OwnerID)
		}
	}
}

func TestGetWidgets_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	widgets, err := repo.GetWidgets(ctx, 0)
	if err != nil {
		t.Fatalf("GetWidgets unknown owner: %v", err)
	}
	if len(widgets) != 0 {
		t.Errorf("expected 0 widgets for unknown owner, got %d", len(widgets))
	}
}

func TestGetWidgets_ConfigIsValidJSON(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	widgets, err := repo.GetWidgets(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	for _, w := range widgets {
		if !json.Valid(w.Config) {
			t.Errorf("widget %d has invalid JSON config: %s", w.ID, string(w.Config))
		}
	}
}

func TestUpsertWidgets_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_901

	widget := &domain.DashboardWidget{
		OwnerID:    isolatedOwner,
		WidgetType: "test_widget",
		PositionX:  0,
		PositionY:  0,
		Width:      2,
		Height:     2,
		Config:     []byte(`{"key":"value"}`),
	}

	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{widget}); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets after upsert: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 widget after upsert, got %d", len(got))
	}
	if got[0].WidgetType != "test_widget" {
		t.Errorf("expected widget_type 'test_widget', got %q", got[0].WidgetType)
	}

	// Cleanup
	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{})
}

func TestUpsertWidgets_ReplacesExisting(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_902

	first := []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "w1", Width: 1, Height: 1, Config: []byte("{}")},
		{OwnerID: isolatedOwner, WidgetType: "w2", Width: 1, Height: 1, Config: []byte("{}")},
	}
	if err := repo.UpsertWidgets(ctx, isolatedOwner, first); err != nil {
		t.Fatalf("UpsertWidgets first: %v", err)
	}

	second := []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "w3", Width: 1, Height: 1, Config: []byte("{}")},
	}
	if err := repo.UpsertWidgets(ctx, isolatedOwner, second); err != nil {
		t.Fatalf("UpsertWidgets second: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets after second upsert: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 widget after replacement, got %d", len(got))
	}
	if got[0].WidgetType != "w3" {
		t.Errorf("expected widget_type 'w3', got %q", got[0].WidgetType)
	}

	// Cleanup
	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{})
}

func TestUpsertWidgets_EmptyListClearsWidgets(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_903

	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "tmp", Width: 1, Height: 1, Config: []byte("{}")},
	}); err != nil {
		t.Fatalf("UpsertWidgets setup: %v", err)
	}

	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}); err != nil {
		t.Fatalf("UpsertWidgets empty: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 widgets after clear, got %d", len(got))
	}
}

func TestGetWidgets_WidgetHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_904
	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "chk", Width: 1, Height: 1, Config: []byte("{}")},
	})
	defer func() {
		_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{})
	}()

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	for _, w := range got {
		if w.ID == 0 {
			t.Error("widget has zero ID after insert")
		}
		if w.CreatedAt.IsZero() {
			t.Error("widget has zero CreatedAt")
		}
	}
}

func TestGetWidgets_OrderedByPosition(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_905
	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "a", PositionX: 1, PositionY: 2, Width: 1, Height: 1, Config: []byte("{}")},
		{OwnerID: isolatedOwner, WidgetType: "b", PositionX: 0, PositionY: 1, Width: 1, Height: 1, Config: []byte("{}")},
		{OwnerID: isolatedOwner, WidgetType: "c", PositionX: 0, PositionY: 0, Width: 1, Height: 1, Config: []byte("{}")},
	})
	defer func() {
		_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{})
	}()

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 widgets, got %d", len(got))
	}
	// Ordered by position_y ASC, position_x ASC: c(0,0), b(0,1), a(1,2)
	if got[0].WidgetType != "c" {
		t.Errorf("first widget should be 'c' (y=0), got %q", got[0].WidgetType)
	}
}
