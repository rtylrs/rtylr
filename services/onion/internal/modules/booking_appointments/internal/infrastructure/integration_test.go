package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_appointments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateAppointment_RejectsOverlappingResource(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	result, err := db.ExecContext(ctx, `
INSERT INTO booking_resources (owner_id, business_id, name, type, capacity, is_active)
VALUES (?, ?, 'Overlap Test Room', 'room', 1, TRUE)`, testutil.OwnerID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}
	resourceID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("resource id: %v", err)
	}
	var firstID int64
	t.Cleanup(func() {
		if firstID != 0 {
			_, _ = db.ExecContext(context.Background(), "DELETE FROM booking_appointments WHERE id = ?", firstID)
		}
		_, _ = db.ExecContext(context.Background(), "DELETE FROM booking_resources WHERE id = ?", resourceID)
	})

	input := baseAppointmentInput()
	input.ResourceID = resourceID
	firstID, err = repo.Create(ctx, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("create first appointment: %v", err)
	}
	_, err = repo.Create(ctx, testutil.BusinessID, input)
	if !errors.Is(err, domain.ErrTimeSlotUnavailable) {
		t.Fatalf("overlapping appointment error = %v, want ErrTimeSlotUnavailable", err)
	}
}

const (
	apptStartTime = "2026-06-01T09:00:00Z"
	apptEndTime   = "2026-06-01T10:00:00Z"
)

func baseAppointmentInput() domain.AppointmentInput {
	return domain.AppointmentInput{
		CustomerName:  "Jane Doe",
		CustomerEmail: "jane@example.com",
		CustomerPhone: "555-1234",
		StartTime:     apptStartTime,
		EndTime:       apptEndTime,
		Status:        "confirmed",
		Notes:         "First visit",
	}
}

func TestListAppointments_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestListAppointments_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestListTodayAppointments_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListToday(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListToday: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestListTodayAppointments_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListToday(ctx, 0)
	if err != nil {
		t.Fatalf("ListToday: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestCreateAppointment_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })
}

func TestCreateAppointment_CustomerNameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseAppointmentInput()
	input.CustomerName = "Alice Smith"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.CustomerName != "Alice Smith" {
				t.Errorf("expected customer name Alice Smith, got %s", a.CustomerName)
			}
			return
		}
	}
	t.Errorf("appointment id=%d not found in list", id)
}

func TestCreateAppointment_CustomerEmailIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseAppointmentInput()
	input.CustomerEmail = "bob@test.com"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.CustomerEmail != "bob@test.com" {
				t.Errorf("expected email bob@test.com, got %s", a.CustomerEmail)
			}
			return
		}
	}
	t.Errorf("appointment id=%d not found in list", id)
}

func TestCreateAppointment_StatusIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseAppointmentInput()
	input.Status = "confirmed"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.Status != "confirmed" {
				t.Errorf("expected status confirmed, got %s", a.Status)
			}
			return
		}
	}
	t.Errorf("appointment id=%d not found in list", id)
}

func TestCreateAppointment_StartTimeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.StartTime.IsZero() {
				t.Errorf("expected non-zero start time")
			}
			return
		}
	}
	t.Errorf("appointment id=%d not found in list", id)
}

func TestUpdateAppointment_ChangesCustomerName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	updated := baseAppointmentInput()
	updated.CustomerName = "Updated Customer"
	if err := repo.Update(ctx, testutil.OwnerID, id, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id && a.CustomerName != "Updated Customer" {
			t.Errorf("expected customer name Updated Customer, got %s", a.CustomerName)
		}
	}
}

func TestUpdateAppointment_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	updated := baseAppointmentInput()
	updated.Status = "cancelled"
	if err := repo.Update(ctx, testutil.OwnerID, id, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id && a.Status != "cancelled" {
			t.Errorf("expected status cancelled, got %s", a.Status)
		}
	}
}

// Regression: an update that changes no column reports RowsAffected()==0 (the
// MySQL driver does not use clientFoundRows), which used to be converted into
// sql.ErrNoRows and surfaced as HTTP 404 even though the locked read in Update
// had already proven the appointment exists.
func TestUpdateAppointment_NoOpUpdateSucceeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseAppointmentInput()
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err := repo.Update(ctx, testutil.OwnerID, id, input); err != nil {
		t.Fatalf("no-op Update = %v, want nil", err)
	}
}

// The locked status read is the existence check, so an unknown appointment id
// must still come back as sql.ErrNoRows for the API to map to 404.
func TestUpdateAppointment_UnknownIDReturnsNoRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.Update(context.Background(), testutil.OwnerID, 999999999, baseAppointmentInput())
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Update unknown id = %v, want sql.ErrNoRows", err)
	}
}

// Regression: sentinel returns skipped the deferred rollback because the local
// err was never assigned, so the failed update abandoned its transaction and
// kept the businesses FOR UPDATE lock until the request context was cancelled.
func TestUpdateAppointment_InvalidStatusReleasesBusinessLock(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	updated := baseAppointmentInput()
	updated.Status = "archived"
	if err := repo.Update(ctx, testutil.OwnerID, id, updated); !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("Update invalid status = %v, want ErrInvalidStatus", err)
	}

	lockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var lockedID int64
	if err := db.QueryRowContext(lockCtx,
		"SELECT id FROM businesses WHERE id = ? FOR UPDATE", testutil.OwnerID).Scan(&lockedID); err != nil {
		t.Fatalf("business %d still locked after failed update: %v", testutil.OwnerID, err)
	}
}

func TestDeleteAppointment_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			t.Errorf("deleted appointment id=%d still in list", id)
		}
	}
}

func TestCreate_ThenList_FindsAppointment(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseAppointmentInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			return
		}
	}
	t.Errorf("appointment id=%d not found after create", id)
}

func TestCreateAppointment_NotesAreStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseAppointmentInput()
	input.Notes = "Allergic to lavender"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.Notes != "Allergic to lavender" {
				t.Errorf("expected notes 'Allergic to lavender', got %s", a.Notes)
			}
			return
		}
	}
	t.Errorf("appointment id=%d not found in list", id)
}
