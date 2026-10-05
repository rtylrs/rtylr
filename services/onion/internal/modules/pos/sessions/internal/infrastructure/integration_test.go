package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/sessions/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/sessions/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSessions_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	sessions, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = sessions
}

func TestSessions_List_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	sessions, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if sessions == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestSessions_List_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	sessions, err := repo.List(context.Background(), 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions for unknown owner, got %d", len(sessions))
	}
}

func TestSessions_List_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	sessions, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, s := range sessions {
		if s.OwnerID != testutil.OwnerID {
			t.Errorf("sessions[%d].OwnerID = %d, want %d", i, s.OwnerID, testutil.OwnerID)
		}
	}
}

func TestSessions_List_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	sessions, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, s := range sessions {
		if s.ID == 0 {
			t.Errorf("sessions[%d].ID is zero", i)
		}
	}
}

func TestSessions_List_AllHaveNonEmptyStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	sessions, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, s := range sessions {
		if s.Status == "" {
			t.Errorf("sessions[%d].Status is empty", i)
		}
	}
}

func TestSessions_Get_FixtureSession3201_StatusOpen(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	session, err := repo.Get(context.Background(), testutil.OwnerID, 3201)
	if err != nil {
		t.Fatalf("Get session 3201: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session for id=3201")
	}
	if session.Status != "open" {
		t.Errorf("session 3201 Status = %q, want 'open'", session.Status)
	}
}

func TestSessions_Get_FixtureSession3201_TotalSales250(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	session, err := repo.Get(context.Background(), testutil.OwnerID, 3201)
	if err != nil {
		t.Fatalf("Get session 3201: %v", err)
	}
	if session.TotalSales != 250 {
		t.Errorf("session 3201 TotalSales = %f, want 250", session.TotalSales)
	}
}

func TestSessions_Get_FixtureSession3201_OpeningCash100(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	session, err := repo.Get(context.Background(), testutil.OwnerID, 3201)
	if err != nil {
		t.Fatalf("Get session 3201: %v", err)
	}
	if session.OpeningFloat != 100 {
		t.Errorf("session 3201 OpeningFloat = %f, want 100", session.OpeningFloat)
	}
}

func TestSessions_Get_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.Get(context.Background(), testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessions_Get_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.Get(context.Background(), 999999, 3201)
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound for wrong owner, got %v", err)
	}
}

func TestSessions_Open_ThenGetReturnsStatusOpen(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locID := int64(3001)
	id, err := repo.Open(ctx, domain.OpenSessionInput{
		OwnerID:      testutil.OwnerID,
		BusinessID:   testutil.BusinessID,
		LocationID:   &locID,
		OpeningFloat: 50.0,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if id == 0 {
		t.Fatal("Open returned ID=0")
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM pos_sessions WHERE id = ?", id); err != nil {
			t.Errorf("cleanup session: %v", err)
		}
	})

	session, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get new session: %v", err)
	}
	if session.Status != "open" {
		t.Errorf("new session Status = %q, want 'open'", session.Status)
	}
}

func TestSessions_Close_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	locID := int64(3001)
	id, err := repo.Open(ctx, domain.OpenSessionInput{
		OwnerID:      testutil.OwnerID,
		BusinessID:   testutil.BusinessID,
		LocationID:   &locID,
		OpeningFloat: 100.0,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	err = repo.Close(ctx, domain.CloseSessionInput{
		ID:           id,
		OwnerID:      testutil.OwnerID,
		BusinessID:   testutil.BusinessID,
		ClosingFloat: 120.0,
	})
	if err != nil {
		t.Fatalf("Close: %v", err)
	}

	session, err := repo.Get(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Get after close: %v", err)
	}
	if session.Status == "open" {
		t.Error("expected session status to change after Close")
	}
}

func TestSessions_Close_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.Close(context.Background(), domain.CloseSessionInput{
		ID:           9999999,
		OwnerID:      testutil.OwnerID,
		BusinessID:   testutil.BusinessID,
		ClosingFloat: 0,
	})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

// Regression: the deferred RELEASE_LOCK ran after tx.Commit(), where
// database/sql returns ErrTxDone without sending the statement, so the MySQL
// named lock stayed on the pooled connection and a later Open for the same
// staff failed with a misleading "already open" after the 5s GET_LOCK timeout.
func TestSessions_Open_ReleasesAdvisoryLock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	staffID := int64(4298)
	id, err := repo.Open(ctx, domain.OpenSessionInput{
		OwnerID:      testutil.OwnerID,
		BusinessID:   testutil.BusinessID,
		StaffID:      &staffID,
		OpeningFloat: 10.0,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM pos_sessions WHERE id = ?`, id)
	})

	assertAdvisoryLockReleased(t, db, fmt.Sprintf("pos_session:%d:%d", testutil.BusinessID, staffID))
}

func assertAdvisoryLockReleased(t *testing.T, db *sql.DB, lockName string) {
	t.Helper()
	var holder sql.NullInt64
	if err := db.QueryRowContext(context.Background(), "SELECT IS_USED_LOCK(?)", lockName).Scan(&holder); err != nil {
		t.Fatalf("check advisory lock %q: %v", lockName, err)
	}
	if holder.Valid {
		t.Fatalf("advisory lock %q is still held by connection %d", lockName, holder.Int64)
	}
}
