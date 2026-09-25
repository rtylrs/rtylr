package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListTables_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Status Filter Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 0, "available")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	for _, tbl := range tables {
		if string(tbl.Status) != "available" {
			t.Errorf("expected status='available', got %q for table id=%d", tbl.Status, tbl.ID)
		}
	}
	_ = id
}

func TestUpdateTable_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Before Update",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	if err := repo.UpdateTable(ctx, testutil.OwnerID, domain.TableInput{
		ID:         id,
		LocationID: 3001,
		Name:       "After Update",
		Capacity:   4,
	}); err != nil {
		t.Fatalf("UpdateTable: %v", err)
	}
}

func TestUpdateTableStatus_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Status Change Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteTable(ctx, testutil.OwnerID, id)
	})

	if err := repo.UpdateTableStatus(ctx, testutil.OwnerID, id, domain.TableStatusOccupied); err != nil {
		t.Fatalf("UpdateTableStatus: %v", err)
	}

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 0, "occupied")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	found := false
	for _, tbl := range tables {
		if tbl.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected table id=%d to appear with status='occupied'", id)
	}
}

func TestDeleteTable_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateTable(ctx, testutil.OwnerID, domain.TableInput{
		LocationID: 3001,
		Name:       "Delete Me Table",
		Capacity:   2,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}

	if err := repo.DeleteTable(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteTable: %v", err)
	}

	tables, err := repo.ListTables(ctx, testutil.OwnerID, 0, "")
	if err != nil {
		t.Fatalf("ListTables: %v", err)
	}
	for _, tbl := range tables {
		if tbl.ID == id {
			t.Errorf("expected table id=%d to be deleted", id)
		}
	}
}

func TestOpenSession_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use a unique staffID to avoid ErrSessionAlreadyOpen from other tests.
	id, err := repo.OpenSession(ctx, testutil.OwnerID, 4199, 3001, 100.00)
	if err != nil {
		t.Fatalf("OpenSession: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive session id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})
}

func TestGetSessionByID_ReturnsFixtureSession(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sess, err := repo.GetSessionByID(ctx, testutil.OwnerID, 3201)
	if err != nil {
		t.Fatalf("GetSessionByID: %v", err)
	}
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
	if sess.ID != 3201 {
		t.Errorf("expected session id=3201, got %d", sess.ID)
	}
}

func TestGetSessionByID_UnknownIDReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetSessionByID(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Error("expected error for unknown session id, got nil")
	}
}

func TestListSessions_ReturnsFixtureSession(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sessions, err := repo.ListSessions(ctx, testutil.OwnerID, 0, "", "")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s.ID == 3201 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected fixture session id=3201 in list")
	}
}

func TestListSessions_FilterByLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sessions, err := repo.ListSessions(ctx, testutil.OwnerID, 3001, "", "")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	for _, s := range sessions {
		if s.LocationID != 3001 {
			t.Errorf("expected locationID=3001, got %d for session id=%d", s.LocationID, s.ID)
		}
	}
}
