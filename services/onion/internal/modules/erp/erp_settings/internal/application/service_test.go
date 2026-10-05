package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/erp_settings/internal/domain"
)

type memoryRepository struct {
	settings    domain.Settings
	getErr      error
	upsertErr   error
	lastInput   domain.SettingsInput
	upsertCalls int
}

func (r *memoryRepository) Get(_ context.Context, _ int64) (domain.Settings, error) {
	return r.settings, r.getErr
}

func (r *memoryRepository) Upsert(_ context.Context, input domain.SettingsInput) error {
	r.upsertCalls++
	r.lastInput = input
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.settings = domain.Settings{
		ID:                 r.settings.ID + 1,
		BusinessID:         input.BusinessID,
		Currency:           input.Currency,
		TaxRate:            input.TaxRate,
		LowStockThreshold:  input.LowStockThreshold,
		DefaultWarehouseID: input.DefaultWarehouseID,
		NotifyOnLowStock:   input.NotifyOnLowStock,
		NotifyOnExpiry:     input.NotifyOnExpiry,
	}
	return nil
}

func TestUpdateSettingsPersistsExplicitFalseFlags(t *testing.T) {
	repo := &memoryRepository{settings: domain.Settings{
		ID:                 9,
		BusinessID:         42,
		Currency:           "USD",
		TaxRate:            8.25,
		LowStockThreshold:  10,
		DefaultWarehouseID: 1,
		NotifyOnLowStock:   true,
		NotifyOnExpiry:     true,
	}}
	service := &Service{repo: repo}

	got, err := service.UpdateSettings(context.Background(), domain.SettingsInput{
		BusinessID:         42,
		Currency:           "EUR",
		TaxRate:            20,
		LowStockThreshold:  4,
		DefaultWarehouseID: 7,
		NotifyOnLowStock:   false,
		NotifyOnExpiry:     false,
	})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", repo.upsertCalls)
	}
	if repo.lastInput.NotifyOnLowStock || repo.lastInput.NotifyOnExpiry {
		t.Fatalf("false notification flags were not passed to repository: %+v", repo.lastInput)
	}
	if got.NotifyOnLowStock || got.NotifyOnExpiry {
		t.Fatalf("false notification flags were not returned: %+v", got)
	}
}

func TestGetSettingsReturnsDefaultsWhenRowDoesNotExist(t *testing.T) {
	repo := &memoryRepository{getErr: domain.ErrSettingsNotFound}
	service := &Service{repo: repo}

	got, err := service.GetSettings(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got.BusinessID != 42 || got.Currency != domain.DefaultCurrency || got.DefaultWarehouseID != domain.DefaultWarehouseID {
		t.Fatalf("unexpected defaults: %+v", got)
	}
}

func TestUpdateSettingsPropagatesRepositoryError(t *testing.T) {
	want := errors.New("write failed")
	service := &Service{repo: &memoryRepository{upsertErr: want}}

	_, err := service.UpdateSettings(context.Background(), domain.SettingsInput{
		BusinessID:         1,
		Currency:           "USD",
		TaxRate:            0,
		LowStockThreshold:  0,
		DefaultWarehouseID: 1,
	})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
