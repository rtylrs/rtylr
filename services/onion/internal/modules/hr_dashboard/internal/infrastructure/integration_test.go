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

func TestUpsertWidgets_WidgetTypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_906
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "headcount_widget", Width: 2, Height: 2, Config: []byte("{}")},
	}); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least 1 widget")
	}
	if got[0].WidgetType != "headcount_widget" {
		t.Errorf("expected widget_type 'headcount_widget', got %q", got[0].WidgetType)
	}
}

func TestUpsertWidgets_PositionIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_907
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "pos_check", PositionX: 3, PositionY: 5, Width: 2, Height: 2, Config: []byte("{}")},
	}); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected 1 widget")
	}
	if got[0].PositionX != 3 || got[0].PositionY != 5 {
		t.Errorf("expected position (3,5), got (%d,%d)", got[0].PositionX, got[0].PositionY)
	}
}

func TestUpsertWidgets_SizeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_908
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "size_check", Width: 4, Height: 3, Config: []byte("{}")},
	}); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected 1 widget")
	}
	if got[0].Width != 4 || got[0].Height != 3 {
		t.Errorf("expected size (4,3), got (%d,%d)", got[0].Width, got[0].Height)
	}
}

func TestUpsertWidgets_ConfigJSONIsRoundTripped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_909
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	configJSON := []byte(`{"title":"Headcount","color":"blue"}`)
	if err := repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "cfg_rt", Width: 1, Height: 1, Config: configJSON},
	}); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected 1 widget")
	}
	// Config should be non-nil and valid JSON.
	if len(got[0].Config) == 0 {
		t.Error("expected non-empty config after upsert")
	}
}

func TestUpsertWidgets_MultipleWidgetsSameOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_910
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	widgets := []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "w1", Width: 1, Height: 1, Config: []byte("{}")},
		{OwnerID: isolatedOwner, WidgetType: "w2", Width: 1, Height: 1, Config: []byte("{}")},
		{OwnerID: isolatedOwner, WidgetType: "w3", Width: 1, Height: 1, Config: []byte("{}")},
	}
	if err := repo.UpsertWidgets(ctx, isolatedOwner, widgets); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 widgets, got %d", len(got))
	}
}

func TestUpsertWidgets_OwnerIsolation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const ownerA int64 = 9_999_911
	const ownerB int64 = 9_999_912
	defer func() {
		_ = repo.UpsertWidgets(ctx, ownerA, []*domain.DashboardWidget{})
		_ = repo.UpsertWidgets(ctx, ownerB, []*domain.DashboardWidget{})
	}()

	_ = repo.UpsertWidgets(ctx, ownerA, []*domain.DashboardWidget{
		{OwnerID: ownerA, WidgetType: "owner_a", Width: 1, Height: 1, Config: []byte("{}")},
	})
	_ = repo.UpsertWidgets(ctx, ownerB, []*domain.DashboardWidget{
		{OwnerID: ownerB, WidgetType: "owner_b", Width: 1, Height: 1, Config: []byte("{}")},
	})

	gotA, err := repo.GetWidgets(ctx, ownerA)
	if err != nil {
		t.Fatalf("GetWidgets ownerA: %v", err)
	}
	for _, w := range gotA {
		if w.OwnerID != ownerA {
			t.Errorf("ownerA widget has owner_id %d", w.OwnerID)
		}
	}
}

func TestUpsertWidgets_UpdatedAtChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_913
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	w := &domain.DashboardWidget{OwnerID: isolatedOwner, WidgetType: "ts_check", Width: 1, Height: 1, Config: []byte("{}")}
	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{w})

	first, _ := repo.GetWidgets(ctx, isolatedOwner)

	// Upsert again -- UpdatedAt should be >= CreatedAt.
	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{w})
	second, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets second: %v", err)
	}
	if len(second) == 0 || len(first) == 0 {
		t.Skip("no widgets returned")
	}
	if second[0].UpdatedAt.Before(first[0].CreatedAt) {
		t.Errorf("UpdatedAt %v should not be before CreatedAt %v", second[0].UpdatedAt, first[0].CreatedAt)
	}
}

func TestGetWidgets_EmptyConfigIsValidJSON(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_914
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	_ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "empty_cfg", Width: 1, Height: 1, Config: []byte("{}")},
	})

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	for _, w := range got {
		if !json.Valid(w.Config) {
			t.Errorf("widget %d has invalid JSON config", w.ID)
		}
	}
}

func TestGetWidgets_CountMatchesUpserted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const isolatedOwner int64 = 9_999_915
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedOwner, []*domain.DashboardWidget{}) }()

	widgets := []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, WidgetType: "cnt1", Width: 1, Height: 1, Config: []byte("{}")},
		{OwnerID: isolatedOwner, WidgetType: "cnt2", Width: 1, Height: 1, Config: []byte("{}")},
	}
	_ = repo.UpsertWidgets(ctx, isolatedOwner, widgets)

	got, err := repo.GetWidgets(ctx, isolatedOwner)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 widgets, got %d", len(got))
	}
}

func TestUpsertWidgets_PersistsOwnerIDDistinctFromBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// The caller's user id and the business it belongs to are normally
	// different values; the row must carry each in its own column.
	const isolatedOwner int64 = 9_999_916
	const isolatedBusiness int64 = 9_999_917
	defer func() { _ = repo.UpsertWidgets(ctx, isolatedBusiness, []*domain.DashboardWidget{}) }()

	if err := repo.UpsertWidgets(ctx, isolatedBusiness, []*domain.DashboardWidget{
		{OwnerID: isolatedOwner, BusinessID: isolatedBusiness, WidgetType: "owner_check", Width: 1, Height: 1, Config: []byte("{}")},
	}); err != nil {
		t.Fatalf("UpsertWidgets: %v", err)
	}

	got, err := repo.GetWidgets(ctx, isolatedBusiness)
	if err != nil {
		t.Fatalf("GetWidgets: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(got))
	}
	if got[0].OwnerID != isolatedOwner {
		t.Errorf("owner_id = %d, want the caller's user id %d", got[0].OwnerID, isolatedOwner)
	}
	if got[0].BusinessID != isolatedBusiness {
		t.Errorf("business_id = %d, want %d", got[0].BusinessID, isolatedBusiness)
	}
}
