package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
)

func TestNewPayload_AllFieldsRoundTrip(t *testing.T) {
	due := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	in := domain.Bill{
		ID:          7,
		OwnerID:     1001,
		SupplierID:  42,
		Amount:      999.99,
		Currency:    "USD",
		DueDate:     due,
		Status:      "open",
		Description: "Office rent",
		CreatedAt:   created,
		UpdatedAt:   created,
	}
	out := newPayload(in)

	if out.ID != in.ID || out.OwnerID != in.OwnerID || out.SupplierID != in.SupplierID {
		t.Errorf("identity didn't round-trip: %+v", out)
	}
	if out.Amount != in.Amount || out.Currency != in.Currency {
		t.Errorf("money didn't round-trip")
	}
	if !out.DueDate.Equal(in.DueDate) || !out.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("dates didn't round-trip")
	}
	if out.Status != in.Status || out.Description != in.Description {
		t.Errorf("status/description didn't round-trip")
	}
}

func TestCreateHandler_Validate_MissingSupplierID(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{SupplierID: 0, Amount: 100, DueDate: "2026-06-01"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing supplier_id")
	}
}

func TestCreateHandler_Validate_InvalidAmount(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{SupplierID: 1, Amount: 0, DueDate: "2026-06-01"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for zero amount")
	}
}

func TestCreateHandler_Validate_MissingDueDate(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{SupplierID: 1, Amount: 100, DueDate: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing due_date")
	}
}

func TestCreateHandler_Validate_BadDueDateFormat(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{SupplierID: 1, Amount: 100, DueDate: "01/06/2026"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for bad date format")
	}
}

func TestCreateHandler_Validate_Valid(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{SupplierID: 1, Amount: 100, DueDate: "2026-06-01"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
