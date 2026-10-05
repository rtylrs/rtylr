package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/booking/reminders/internal/domain"
	"github.com/voxire/onion/internal/modules/booking/reminders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const reminderScheduledAt = "2026-06-01T08:00:00Z"

func TestListReminders_ZeroOwner_ReturnsEmpty(t *testing.T) {
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

func TestListReminders_IsNotNil(t *testing.T) {
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

func TestCreateReminder_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateReminder_AppointmentIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// AppointmentID=0 means nullable FK is NULL
	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		AppointmentID: 0,
		Type:          "email",
		ScheduledAt:   reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			// AppointmentID=0 when FK is null
			if r.AppointmentID != 0 {
				t.Errorf("expected AppointmentID=0, got %d", r.AppointmentID)
			}
			return
		}
	}
	t.Errorf("reminder id=%d not found in list", id)
}

func TestCreateReminder_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "sms",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.Type != "sms" {
				t.Errorf("expected type sms, got %s", r.Type)
			}
			return
		}
	}
	t.Errorf("reminder id=%d not found in list", id)
}

func TestCreateReminder_ScheduledAtIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.ScheduledAt.IsZero() {
				t.Errorf("expected non-zero ScheduledAt")
			}
			return
		}
	}
	t.Errorf("reminder id=%d not found in list", id)
}

func TestCreate_ThenList_FindsReminder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			return
		}
	}
	t.Errorf("reminder id=%d not found after create", id)
}

func TestListReminders_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateReminder_TypeEmail_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create email reminder: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.Type != "email" {
				t.Errorf("expected type email, got %s", r.Type)
			}
			return
		}
	}
	t.Errorf("email reminder id=%d not found", id)
}

func TestCreateReminder_TypeSMS_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "sms",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create sms reminder: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.Type != "sms" {
				t.Errorf("expected type sms, got %s", r.Type)
			}
			return
		}
	}
	t.Errorf("sms reminder id=%d not found", id)
}

func TestListReminders_MultipleOwners_Scoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create for owner 1001
	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: reminderScheduledAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Owner 9999999 should not see owner 1001's reminders
	other, err := repo.List(ctx, 9999999)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, r := range other {
		if r.ID == id {
			t.Errorf("reminder id=%d leaked to another owner", id)
		}
	}
}

func TestCreateReminder_ScheduledAtFutureDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReminderInput{
		Type:        "email",
		ScheduledAt: "2027-01-15T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("Create future reminder: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.ScheduledAt.IsZero() {
				t.Errorf("expected non-zero ScheduledAt for future date")
			}
			return
		}
	}
	t.Errorf("reminder id=%d not found in list", id)
}
