package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_users/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Fixture pos_users: 3601 Cashier A (cashier), 3602 Server B (server), location_id=3001.
const fixtureLocationID int64 = 3001

func TestPOSUsers_ListByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByOwner(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	_ = users
}

func TestPOSUsers_ListByOwner_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByOwner(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	if users == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestPOSUsers_ListByOwner_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByOwner(context.Background(), 999999)
	if err != nil {
		t.Fatalf("ListByOwner wrong owner: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users for unknown owner, got %d", len(users))
	}
}

func TestPOSUsers_ListByLocation_ContainsCashierA(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByLocation(context.Background(), testutil.OwnerID, fixtureLocationID)
	if err != nil {
		t.Fatalf("ListByLocation: %v", err)
	}

	found := false
	for _, u := range users {
		if u.ID == 3601 && u.Name == "Cashier A" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected user id=3601 'Cashier A' in location 3001, got %+v", users)
	}
}

func TestPOSUsers_ListByLocation_ContainsServerB(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByLocation(context.Background(), testutil.OwnerID, fixtureLocationID)
	if err != nil {
		t.Fatalf("ListByLocation: %v", err)
	}

	found := false
	for _, u := range users {
		if u.ID == 3602 && u.Name == "Server B" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected user id=3602 'Server B' in location 3001, got %+v", users)
	}
}

func TestPOSUsers_ListByLocation_CashierAHasRoleCashier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByLocation(context.Background(), testutil.OwnerID, fixtureLocationID)
	if err != nil {
		t.Fatalf("ListByLocation: %v", err)
	}
	for _, u := range users {
		if u.ID == 3601 {
			if u.Role != domain.RoleCashier {
				t.Errorf("user 3601 Role = %q, want %q", u.Role, domain.RoleCashier)
			}
			return
		}
	}
	t.Error("user 3601 not found")
}

func TestPOSUsers_ListByLocation_ServerBHasRoleServer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByLocation(context.Background(), testutil.OwnerID, fixtureLocationID)
	if err != nil {
		t.Fatalf("ListByLocation: %v", err)
	}
	for _, u := range users {
		if u.ID == 3602 {
			if u.Role != domain.RoleServer {
				t.Errorf("user 3602 Role = %q, want %q", u.Role, domain.RoleServer)
			}
			return
		}
	}
	t.Error("user 3602 not found")
}

func TestPOSUsers_ListByLocation_WrongLocationReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByLocation(context.Background(), testutil.OwnerID, 999999)
	if err != nil {
		t.Fatalf("ListByLocation wrong location: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users for unknown location, got %d", len(users))
	}
}

func TestPOSUsers_ListByLocation_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByLocation(context.Background(), testutil.OwnerID, fixtureLocationID)
	if err != nil {
		t.Fatalf("ListByLocation: %v", err)
	}
	for i, u := range users {
		if u.ID == 0 {
			t.Errorf("users[%d].ID is zero", i)
		}
	}
}

func TestPOSUsers_Create_ThenListByOwnerContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreatePOSUserInput{
		UserID:     90001,
		LocationID: fixtureLocationID,
		Name:       "Integration Tester",
		PIN:        "9999",
		Role:       domain.RoleCashier,
	}, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned ID=0")
	}

	users, err := repo.ListByOwner(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	found := false
	for _, u := range users {
		if u.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("new user id=%d not found in list", id)
	}
}

func TestPOSUsers_FindForVerify_UnknownReturnsErrUserNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.FindForVerify(context.Background(), testutil.OwnerID, fixtureLocationID, 999999)
	if !errors.Is(err, infrastructure.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPOSUsers_Deactivate_IsCallable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreatePOSUserInput{
		UserID:     90002,
		LocationID: fixtureLocationID,
		Name:       "To Deactivate",
		PIN:        "1111",
		Role:       domain.RoleCashier,
	}, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Deactivate(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
}

func TestPOSUsers_UpdatePINHash_RefusesEmptyHash(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.UpdatePINHash(context.Background(), testutil.OwnerID, 3601, "")
	if err == nil {
		t.Fatal("expected error for empty pin hash, got nil")
	}
}

func TestPOSUsers_ListByOwner_AllHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	users, err := repo.ListByOwner(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	for i, u := range users {
		if u.Name == "" {
			t.Errorf("users[%d].Name is empty", i)
		}
	}
}
