package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
)

func TestNewPayload_RoundTripsAllFields(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	in := domain.BankAccount{
		ID:           1,
		OwnerID:      1001,
		Name:         "Operating",
		Institution:  "Bank of Beirut",
		AccountType:  "checking",
		Balance:      12345.67,
		Currency:     "USD",
		IsConnected:  true,
		LastSyncedAt: now,
		CreatedAt:    now.Add(-24 * time.Hour),
	}

	out := newPayload(in)

	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.Name != in.Name {
		t.Errorf("identity fields didn't round-trip: %+v", out)
	}
	if out.Institution != in.Institution || out.AccountType != in.AccountType {
		t.Errorf("institution/account_type didn't round-trip: %+v", out)
	}
	if out.Balance != in.Balance || out.Currency != in.Currency {
		t.Errorf("money fields didn't round-trip: %+v", out)
	}
	if out.IsConnected != in.IsConnected {
		t.Errorf("is_connected didn't round-trip")
	}
	if !out.LastSyncedAt.Equal(in.LastSyncedAt) || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("timestamps didn't round-trip")
	}
}

func TestNewPayload_HandlesZeroValue(t *testing.T) {
	out := newPayload(domain.BankAccount{})
	if out.ID != 0 || out.Name != "" || !out.LastSyncedAt.IsZero() {
		t.Errorf("zero value should pass through cleanly: %+v", out)
	}
}

func TestCreateHandler_Validate_MissingName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: ""}
	if err := h.Validate(nil, req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "   "}
	if err := h.Validate(nil, req); err == nil {
		t.Error("expected error for whitespace-only name")
	}
}

func TestCreateHandler_Validate_Valid(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Operating"}
	if err := h.Validate(nil, req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
