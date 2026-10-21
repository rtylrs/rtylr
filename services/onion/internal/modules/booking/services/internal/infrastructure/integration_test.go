package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/booking/services/internal/domain"
	"github.com/voxire/onion/internal/modules/booking/services/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestListBookingServices_ZeroOwner_ReturnsEmpty(t *testing.T) {
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

func TestListBookingServices_IsNotNil(t *testing.T) {
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

func TestCreateBookingService_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "Haircut",
		DurationMinutes: 30,
		Price:           money.Amount(2500),
		IsActive:        true,
		Color:           "#ff0000",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })
}

func TestCreateBookingService_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "Massage",
		DurationMinutes: 60,
		Price:           money.Amount(8000),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.Name != "Massage" {
				t.Errorf("expected name Massage, got %s", s.Name)
			}
			return
		}
	}
	t.Errorf("created service id=%d not found in list", id)
}

func TestCreateBookingService_DurationMinutesIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "Facial",
		DurationMinutes: 45,
		Price:           money.Amount(6000),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.DurationMinutes != 45 {
				t.Errorf("expected duration 45, got %d", s.DurationMinutes)
			}
			return
		}
	}
	t.Errorf("created service id=%d not found in list", id)
}

func TestCreateBookingService_PriceIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "Pedicure",
		DurationMinutes: 50,
		Price:           money.Amount(4550),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.Price != money.Amount(4550) {
				t.Errorf("expected price 45.50, got %s", s.Price.String())
			}
			return
		}
	}
	t.Errorf("created service id=%d not found in list", id)
}

func TestCreateBookingService_IsActiveIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "Waxing",
		DurationMinutes: 20,
		Price:           money.Amount(3000),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if !s.IsActive {
				t.Errorf("expected IsActive=true, got false")
			}
			return
		}
	}
	t.Errorf("created service id=%d not found in list", id)
}

func TestCreateBookingService_ColorIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "Manicure",
		DurationMinutes: 30,
		Price:           money.Amount(2500),
		Color:           "#00ff00",
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			if s.Color != "#00ff00" {
				t.Errorf("expected color #00ff00, got %s", s.Color)
			}
			return
		}
	}
	t.Errorf("created service id=%d not found in list", id)
}

func TestUpdateBookingService_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "OldServiceName",
		DurationMinutes: 30,
		Price:           money.Amount(2000),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BookingServiceInput{
		Name:            "NewServiceName",
		DurationMinutes: 30,
		Price:           money.Amount(2000),
		IsActive:        true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id && s.Name != "NewServiceName" {
			t.Errorf("expected name NewServiceName, got %s", s.Name)
		}
	}
}

func TestUpdateBookingService_ChangesPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "PriceService",
		DurationMinutes: 30,
		Price:           money.Amount(1000),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.BookingServiceInput{
		Name:            "PriceService",
		DurationMinutes: 30,
		Price:           money.Amount(9999),
		IsActive:        true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id && s.Price != money.Amount(9999) {
			t.Errorf("expected price 99.99, got %s", s.Price.String())
		}
	}
}

func TestDeleteBookingService_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "ToDelete",
		DurationMinutes: 15,
		Price:           money.Amount(500),
		IsActive:        true,
	})
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
	for _, s := range items {
		if s.ID == id {
			t.Errorf("deleted service id=%d still in list", id)
		}
	}
}

func TestDeleteBookingService_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.OwnerID, 99999999); err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestCreate_ThenList_FindsService(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "FindMe",
		DurationMinutes: 60,
		Price:           money.Amount(7500),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range items {
		if s.ID == id {
			return
		}
	}
	t.Errorf("service id=%d not found after create", id)
}

func TestListBookingServices_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, domain.BookingServiceInput{
		Name:            "CountIncrease",
		DurationMinutes: 30,
		Price:           money.Amount(1500),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}
